// Package impact — слайс slice-03-diff-breaking-change: оценка влияния новой
// версии спеки поставщика на живущих потребителей (oasdiff + граф).
package impact

import (
	"errors"

	"pinout-netlist/internal/shared/specloader"
	"pinout-netlist/internal/shared/store"
)

// Сентинелы среза (HTTP-маппинг — adapter.go).
var (
	// ErrBadRequest — запрос не соответствует схеме (нет provider/источников,
	// источник не exactly-one).
	ErrBadRequest = errors.New("BAD_REQUEST")
)

// ImpactRequest — единственный вход среза (тело POST /impact).
type ImpactRequest struct {
	Body []byte
}

// impactEnvelope — wire-форма тела (openapi.yaml ImpactRequest).
type impactEnvelope struct {
	Provider  string                `json:"provider"`
	ToVersion string                `json:"to_version"`
	FromSpec  specloader.SpecSource `json:"from_spec"`
	ToSpec    specloader.SpecSource `json:"to_spec"`
}

// ImpactInput — декодированный и проверенный запрос.
type ImpactInput struct {
	Provider  string
	ToVersion string
	FromSpec  specloader.SpecSource
	ToSpec    specloader.SpecSource
}

// AffectedSubject — затронутый субъект (METHOD /path) с кодом изменения.
type AffectedSubject struct {
	Code    string
	Message string
	Subject string
}

// BreakingChange — единственный результат (200; openapi.yaml BreakingChange).
type BreakingChange struct {
	Provider          string              `json:"provider"`
	FromVersion       string              `json:"from_version"`
	ToVersion         string              `json:"to_version"`
	AffectedConsumers []string            `json:"affected_consumers"`
	Details           []store.ReportError `json:"details"`
}
