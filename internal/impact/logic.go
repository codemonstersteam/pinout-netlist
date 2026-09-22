// logic.go — чистая логика среза: диф версий спеки (oasdiff, honest reuse) и
// вывод задетых потребителей. contracts.md §DiffSpecs/§AffectedConsumers.
package impact

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
	"github.com/oasdiff/oasdiff/load"

	"pinout-netlist/internal/shared/store"
)

// DiffSpecs — breaking-диф старой→новой спеки: oasdiff checker (honest reuse:
// операции/параметры/типы) + пиновая проходка схем (удаление поля ответа — ядро
// R3 валидаторов, которого нет в политике oasdiff; новое обязательное поле
// запроса — R2). Каждый breaking change → субъект `METHOD /path`.
func DiffSpecs(from, to *load.SpecInfo) []AffectedSubject {
	var out []AffectedSubject
	d, osm, err := diff.GetWithOperationsSourcesMap(&diff.Config{}, from, to)
	if err == nil && d != nil {
		config := checker.NewConfig(checker.GetAllChecks())
		localizer := checker.NewLocalizer("en")
		for _, c := range checker.CheckBackwardCompatibility(config, d, osm) {
			if !c.IsBreaking() {
				continue
			}
			out = append(out, AffectedSubject{
				Code:    c.GetId(),
				Message: c.GetUncolorizedText(localizer),
				Subject: subjectOf(c),
			})
		}
	}
	out = append(out, schemaBreakingSubjects(from.Spec, to.Spec)...)
	return out
}

// subjectOf — канонный субъект `METHOD /path` из изменения oasdiff.
func subjectOf(c checker.Change) string {
	method := strings.ToUpper(strings.TrimSpace(c.GetOperation()))
	if path := strings.TrimSpace(c.GetPath()); path != "" {
		return method + " " + path
	}
	return method
}

// schemaBreakingSubjects — пиновые правила на плоских полях (как DeriveProvider
// Operation валидаторов): (1) поле ответа удалено (R3 — потребитель его читает);
// (2) в запросе появилось новое обязательное поле (R2 — потребитель его не шлёт).
func schemaBreakingSubjects(from, to *openapi3.T) []AffectedSubject {
	var out []AffectedSubject
	for path, fromItem := range from.Paths.Map() {
		for method, fromOp := range fromItem.Operations() {
			subject := strings.ToUpper(method) + " " + path
			toItem := to.Paths.Find(path)
			if toItem == nil {
				continue // удаление операции ловит oasdiff
			}
			toOp := toItem.Operations()[method]
			if toOp == nil {
				continue
			}

			// R3: удалённые поля тела ответа 2xx (top-level, плоско).
			for field := range responseFields(fromOp) {
				if _, still := responseFields(toOp)[field]; !still {
					out = append(out, AffectedSubject{
						Code:    "response-field-removed",
						Message: fmt.Sprintf("response field %q of %s was removed", field, subject),
						Subject: subject,
					})
				}
			}

			// R2: новое обязательное поле запроса.
			fromReq := requiredRequestFields(fromOp)
			for field := range requiredRequestFields(toOp) {
				if _, existed := fromReq[field]; !existed {
					out = append(out, AffectedSubject{
						Code:    "request-required-field-added",
						Message: fmt.Sprintf("request field %q of %s became required", field, subject),
						Subject: subject,
					})
				}
			}
		}
	}
	return out
}

// responseFields — плоская карта top-level полей JSON-тела первого 2xx-ответа.
func responseFields(op *openapi3.Operation) map[string]struct{} {
	out := map[string]struct{}{}
	if op == nil || op.Responses == nil {
		return out
	}
	for status, resp := range op.Responses.Map() {
		n, _ := strconv.Atoi(status)
		if n < 200 || n > 299 || resp.Value == nil {
			continue
		}
		media, ok := resp.Value.Content["application/json"]
		if !ok || media.Schema == nil || media.Schema.Value == nil {
			continue
		}
		for name := range media.Schema.Value.Properties {
			out[name] = struct{}{}
		}
	}
	return out
}

// requiredRequestFields — обязательные поля тела запроса (плоско).
func requiredRequestFields(op *openapi3.Operation) map[string]struct{} {
	out := map[string]struct{}{}
	if op == nil || op.RequestBody == nil || op.RequestBody.Value == nil {
		return out
	}
	media, ok := op.RequestBody.Value.Content["application/json"]
	if !ok || media.Schema == nil || media.Schema.Value == nil {
		return out
	}
	for _, name := range media.Schema.Value.Required {
		out[name] = struct{}{}
	}
	return out
}

// AffectedConsumers — уникальные живущие потребители, чьи рёбра (Subject) задеты
// дифом; сортировка не нужна — потребитель один раз (map-дедуп, порядок вставки).
func AffectedConsumers(subjects []AffectedSubject, edges []store.Edge) []string {
	if len(subjects) == 0 {
		return []string{}
	}
	hit := make(map[string]struct{}, len(subjects))
	for _, s := range subjects {
		hit[s.Subject] = struct{}{}
	}
	seen := make(map[string]struct{})
	var out []string
	for _, e := range edges {
		if _, affected := hit[e.Subject]; !affected {
			continue
		}
		if _, dup := seen[e.Consumer]; dup {
			continue
		}
		seen[e.Consumer] = struct{}{}
		out = append(out, e.Consumer)
	}
	if out == nil {
		return []string{}
	}
	return out
}

// detailsOf — детали по каждому затронутому субъекту (форма ValidationError).
func detailsOf(subjects []AffectedSubject) []store.ReportError {
	out := make([]store.ReportError, 0, len(subjects))
	for _, s := range subjects {
		out = append(out, store.ReportError{Code: s.Code, Message: s.Message, Subject: s.Subject})
	}
	return out
}
