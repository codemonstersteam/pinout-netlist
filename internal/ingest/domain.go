// Package ingest — слайс slice-01-ingest-report: приём отчёта валидатора (канон 1.1),
// вывод рёбер по субъектам, обновление графа и истории версий.
// Доменные типы + конструкторы (valid by construction) + чистая логика;
// contracts.md §ParseReport/§SubjectsOf, module-tree.md.
package ingest

import "errors"

// Сентинелы среза (маппинг на HTTP — в adapter.go).
var (
	// ErrBadReport — тело не является каноническим отчётом 1.1.
	ErrBadReport = errors.New("BAD_REPORT")
	// ErrUnsupportedSchemaVersion — schema_version ≠ "1.1".
	ErrUnsupportedSchemaVersion = errors.New("UNSUPPORTED_SCHEMA_VERSION")
)

// RawError — элемент errors[] отчёта (wire-форма, канон 1.1).
type RawError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Subject string `json:"subject"`
}

// RawReport — wire-форма отчёта канона 1.1 (ровно то, что читает netlist).
type RawReport struct {
	SchemaVersion string `json:"schema_version"`
	Validator     string `json:"validator"`
	Interaction   string `json:"interaction"`
	Consumer      struct {
		Name string `json:"name"`
	} `json:"consumer"`
	GeneratedAt string `json:"generated_at"`
	Compatible  bool   `json:"compatible"`
	Provenance  struct {
		Provider        string `json:"provider"`
		ProviderVersion string `json:"provider_version"`
		CapturedHash    string `json:"captured_hash"`
	} `json:"provenance"`
	Errors              []RawError `json:"errors"`
	UncoveredOperations []string   `json:"uncovered_operations"`
	UncoveredChannels   []string   `json:"uncovered_channels"`
}

// Report — валидный отчёт канона 1.1 (следствие NewReport). Идентичность сторон
// и provenance проверены; субъекты ошибок извлечены.
type Report struct {
	schemaVersion   string
	validator       string
	interaction     string
	consumerName    string
	generatedAt     string
	compatible      bool
	provider        string
	providerVersion string
	capturedHash    string
	errors          []RawError
	uncovered       []string
}

// NewReport — конструктор канона 1.1: schema_version ровно "1.1"; идентичность
// (validator/interaction/consumer.name/provenance) непуста; инвариант
// compatible ⇔ errors == []. Отчёты exit 2/3 (без provenance/идентичности)
// сюда не проходят — ErrBadReport.
func NewReport(raw RawReport) (Report, error) {
	if raw.SchemaVersion != "1.1" {
		return Report{}, ErrUnsupportedSchemaVersion
	}
	if raw.Validator == "" || raw.Interaction == "" || raw.Consumer.Name == "" ||
		raw.Provenance.Provider == "" || raw.Provenance.CapturedHash == "" {
		return Report{}, ErrBadReport
	}
	if raw.GeneratedAt == "" {
		return Report{}, ErrBadReport
	}
	if raw.Compatible != (len(raw.Errors) == 0) {
		return Report{}, ErrBadReport
	}
	uncovered := raw.UncoveredOperations
	if raw.UncoveredChannels != nil {
		uncovered = append(append([]string{}, raw.UncoveredOperations...), raw.UncoveredChannels...)
	}
	return Report{
		schemaVersion:   raw.SchemaVersion,
		validator:       raw.Validator,
		interaction:     raw.Interaction,
		consumerName:    raw.Consumer.Name,
		generatedAt:     raw.GeneratedAt,
		compatible:      raw.Compatible,
		provider:        raw.Provenance.Provider,
		providerVersion: raw.Provenance.ProviderVersion,
		capturedHash:    raw.Provenance.CapturedHash,
		errors:          raw.Errors,
		uncovered:       uncovered,
	}, nil
}

// IngestInput — декодированный конверт {report, subjects?} + валидный отчёт.
type IngestInput struct {
	Report   Report
	Subjects []string
}
