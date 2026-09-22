// constructors.go — valid-by-construction конструкторы среза: ребро и вердикт
// собираются ТОЛЬКО здесь, каждое поле проверяется (module-tree.md; step-03).
package ingest

import (
	"errors"

	"pinout-netlist/internal/shared/store"
)

// ErrEmptySubject — субъект ребра пуст (проверка NewEdge).
var ErrEmptySubject = errors.New("EMPTY_SUBJECT")

// NewEdge — ребро из валидного отчёта и субъекта: идентичность сторон и
// interaction уже доказаны NewReport; здесь проверяется сам субъект.
func NewEdge(r Report, subject string) (store.Edge, error) {
	if subject == "" {
		return store.Edge{}, ErrEmptySubject
	}
	return store.Edge{
		Consumer:    r.consumerName,
		Provider:    r.provider,
		Subject:     subject,
		Interaction: r.interaction,
	}, nil
}

// NewVerdictRecord — вердикт ребра из отчёта: совместимость = «нет ошибки с
// этим субъектом» (агрегатный compatible разложен; messages.md §1), ошибки —
// только этого субъекта; провенанс и время — сквозные эхи отчёта.
func NewVerdictRecord(r Report, subject string) (store.VerdictRecord, error) {
	edge, err := NewEdge(r, subject)
	if err != nil {
		return store.VerdictRecord{}, err
	}
	rerrs := ErrorsOfSubject(r, subject)
	serrs := make([]store.ReportError, 0, len(rerrs))
	for _, e := range rerrs {
		serrs = append(serrs, store.ReportError{Code: e.Code, Message: e.Message, Subject: e.Subject})
	}
	return store.VerdictRecord{
		Edge:            edge,
		Compatible:      EdgeCompatible(r, subject),
		Errors:          serrs,
		ProviderVersion: r.providerVersion,
		CapturedHash:    r.capturedHash,
		At:              r.generatedAt,
	}, nil
}
