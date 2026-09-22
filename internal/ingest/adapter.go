// adapter.go — ingress-адаптер среза: JSON-декод конверта {report, subjects?} и
// маппинг сентинел → HTTP-статус (единственное место маппинга; contracts.md
// §ParseIngestBody, «Error model»).
package ingest

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ingestEnvelope — wire-форма тела POST /reports (openapi.yaml IngestReportRequest).
type ingestEnvelope struct {
	Report   RawReport `json:"report"`
	Subjects []string  `json:"subjects"`
}

// ParseIngestBody — декод тела → IngestInput (валидный отчёт — NewReport).
// Ошибки: ErrBadReport (битый JSON/неканонический отчёт), ErrUnsupportedSchemaVersion.
func ParseIngestBody(body []byte) (IngestInput, error) {
	var env ingestEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return IngestInput{}, ErrBadReport
	}
	rep, err := NewReport(env.Report)
	if err != nil {
		return IngestInput{}, err
	}
	return IngestInput{Report: rep, Subjects: env.Subjects}, nil
}

// HTTPStatus — маппинг сентинела среза на статус ответа (adapter-таблица).
func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrUnsupportedSchemaVersion):
		return http.StatusConflict // 409
	case errors.Is(err, ErrBadReport):
		return http.StatusBadRequest // 400
	default:
		return http.StatusInternalServerError // 500 (хранилище и пр.)
	}
}
