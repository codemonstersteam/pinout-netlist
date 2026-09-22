package ingest

import (
	"errors"
	"testing"
)

// rawOK — минимальный канонический отчёт 1.1 для конструирования ветвей.
func rawOK() RawReport {
	var r RawReport
	r.SchemaVersion = "1.1"
	r.Validator = "pinout-openapi"
	r.Interaction = "sync"
	r.Consumer.Name = "wallet-ui"
	r.GeneratedAt = "2026-09-22T12:00:00Z"
	r.Compatible = true
	r.Provenance.Provider = "balance-service"
	r.Provenance.ProviderVersion = "1.0.0"
	r.Provenance.CapturedHash = "sha256:aa"
	return r
}

// TestNewReport_Happy — 1 happy: валидный отчёт проходит, поля переносятся.
func TestNewReport_Happy(t *testing.T) {
	raw := rawOK()
	raw.UncoveredOperations = []string{"GET /transfers"}
	rep, err := NewReport(raw)
	if err != nil {
		t.Fatalf("NewReport() error: %v", err)
	}
	if rep.consumerName != "wallet-ui" || rep.provider != "balance-service" || rep.interaction != "sync" {
		t.Errorf("identity fields mismatch: %+v", rep)
	}
	if len(rep.uncovered) != 1 || rep.uncovered[0] != "GET /transfers" {
		t.Errorf("Uncovered = %v", rep.uncovered)
	}
}

// TestNewReport_UnsupportedSchemaVersion — ветвь: schema_version ≠ "1.1" → 409-сентинел.
func TestNewReport_UnsupportedSchemaVersion(t *testing.T) {
	raw := rawOK()
	raw.SchemaVersion = "1.0"
	if _, err := NewReport(raw); !errors.Is(err, ErrUnsupportedSchemaVersion) {
		t.Errorf("err = %v, want ErrUnsupportedSchemaVersion", err)
	}
}

// TestNewReport_MissingIdentity — ветвь: нет consumer.name → BAD_REPORT.
func TestNewReport_MissingIdentity(t *testing.T) {
	raw := rawOK()
	raw.Consumer.Name = ""
	if _, err := NewReport(raw); !errors.Is(err, ErrBadReport) {
		t.Errorf("err = %v, want ErrBadReport", err)
	}
}

// TestNewReport_MissingProvenance — ветвь: exit-3-конверт без provenance → BAD_REPORT.
func TestNewReport_MissingProvenance(t *testing.T) {
	raw := rawOK()
	raw.Provenance.CapturedHash = ""
	if _, err := NewReport(raw); !errors.Is(err, ErrBadReport) {
		t.Errorf("err = %v, want ErrBadReport", err)
	}
}

// TestNewReport_MissingGeneratedAt — ветвь: нет generated_at → BAD_REPORT.
func TestNewReport_MissingGeneratedAt(t *testing.T) {
	raw := rawOK()
	raw.GeneratedAt = ""
	if _, err := NewReport(raw); !errors.Is(err, ErrBadReport) {
		t.Errorf("err = %v, want ErrBadReport", err)
	}
}

// TestNewReport_InvariantViolation — ветвь: compatible=true при errors≠[] → BAD_REPORT.
func TestNewReport_InvariantViolation(t *testing.T) {
	raw := rawOK()
	raw.Compatible = false // и пустые ошибки — обратное нарушение того же инварианта
	if _, err := NewReport(raw); !errors.Is(err, ErrBadReport) {
		t.Errorf("err = %v, want ErrBadReport", err)
	}
}

// TestSubjectsOf_Happy — 1 happy: объединение всех трёх источников.
func TestSubjectsOf_Happy(t *testing.T) {
	rep := Report{errors: []RawError{{Subject: "GET /a"}}, uncovered: []string{"POST /b"}}
	got := SubjectsOf(rep, []string{"GET /c"})
	want := []string{"GET /c", "GET /a", "POST /b"}
	if len(got) != len(want) {
		t.Fatalf("SubjectsOf = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SubjectsOf[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestSubjectsOf_OnlySubjects / OnlyErrorSubjects / OnlyUncovered — ветви источников.
func TestSubjectsOf_OnlySubjects(t *testing.T) {
	if got := SubjectsOf(Report{}, []string{"GET /x"}); len(got) != 1 || got[0] != "GET /x" {
		t.Errorf("SubjectsOf = %v", got)
	}
}

func TestSubjectsOf_OnlyErrorSubjects(t *testing.T) {
	rep := Report{errors: []RawError{{Subject: "GET /e"}}}
	if got := SubjectsOf(rep, nil); len(got) != 1 || got[0] != "GET /e" {
		t.Errorf("SubjectsOf = %v", got)
	}
}

func TestSubjectsOf_OnlyUncovered(t *testing.T) {
	rep := Report{uncovered: []string{"orders.v1 receive OrderCreated"}}
	if got := SubjectsOf(rep, nil); len(got) != 1 {
		t.Errorf("SubjectsOf = %v", got)
	}
}

// TestSubjectsOf_Dedup — ветвь: дубликат схлопывается.
func TestSubjectsOf_Dedup(t *testing.T) {
	rep := Report{errors: []RawError{{Subject: "GET /a"}, {Subject: "GET /a"}}}
	if got := SubjectsOf(rep, []string{"GET /a"}); len(got) != 1 {
		t.Errorf("SubjectsOf = %v, want deduplicated single", got)
	}
}

// TestNewEdge_Happy — 1 happy: ребро из валидного отчёта.
func TestNewEdge_Happy(t *testing.T) {
	rep, _ := NewReport(rawOK())
	e, err := NewEdge(rep, "GET /accounts/{id}")
	if err != nil {
		t.Fatalf("NewEdge() error: %v", err)
	}
	if e.Consumer != "wallet-ui" || e.Provider != "balance-service" || e.Subject != "GET /accounts/{id}" || e.Interaction != "sync" {
		t.Errorf("edge = %+v", e)
	}
}

// TestNewEdge_EmptySubject — ветвь: пустой субъект.
func TestNewEdge_EmptySubject(t *testing.T) {
	rep, _ := NewReport(rawOK())
	if _, err := NewEdge(rep, ""); !errors.Is(err, ErrEmptySubject) {
		t.Errorf("err = %v, want ErrEmptySubject", err)
	}
}

// TestNewVerdictRecord_Happy — 1 happy: зелёное ребро у совместимого отчёта.
func TestNewVerdictRecord_Happy(t *testing.T) {
	rep, _ := NewReport(rawOK())
	v, err := NewVerdictRecord(rep, "GET /accounts/{id}")
	if err != nil {
		t.Fatalf("NewVerdictRecord() error: %v", err)
	}
	if !v.Compatible || len(v.Errors) != 0 || v.At != "2026-09-22T12:00:00Z" {
		t.Errorf("verdict = %+v", v)
	}
}

// TestNewVerdictRecord_OtherSubjectsError — ветвь: ошибка чужого субъекта не красит ребро
// (агрегатный compatible разложен по субъектам — ловушка «все рёбра красные»).
func TestNewVerdictRecord_OtherSubjectsError(t *testing.T) {
	raw := rawOK()
	raw.Compatible = false
	raw.Errors = []RawError{{Code: "OP_NOT_IN_PROVIDER", Subject: "POST /other"}}
	rep, _ := NewReport(raw)
	v, err := NewVerdictRecord(rep, "GET /accounts/{id}")
	if err != nil {
		t.Fatalf("NewVerdictRecord() error: %v", err)
	}
	if !v.Compatible {
		t.Errorf("edge must stay compatible when another subject failed: %+v", v)
	}
	broken, _ := NewVerdictRecord(rep, "POST /other")
	if broken.Compatible || len(broken.Errors) != 1 {
		t.Errorf("broken edge = %+v", broken)
	}
}
