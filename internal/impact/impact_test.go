package impact

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pinout-netlist/internal/shared/specloader"
	"pinout-netlist/internal/shared/store"
)

// ── NewSpecSource (3) ──────────────────────────────────────────────────────────

func TestNewSpecSource_Happy(t *testing.T) {
	src, err := specloader.NewSpecSource("/tmp/s.yaml", "")
	if err != nil || src.SpecPath != "/tmp/s.yaml" || src.SpecURL != "" {
		t.Errorf("NewSpecSource = %+v, %v", src, err)
	}
}

func TestNewSpecSource_BothSet(t *testing.T) {
	if _, err := specloader.NewSpecSource("/a", "http://b"); !errors.Is(err, specloader.ErrBadRequestSource) {
		t.Errorf("err = %v, want ErrBadRequestSource", err)
	}
}

func TestNewSpecSource_NeitherSet(t *testing.T) {
	if _, err := specloader.NewSpecSource("", ""); !errors.Is(err, specloader.ErrBadRequestSource) {
		t.Errorf("err = %v, want ErrBadRequestSource", err)
	}
}

// ── DiffSpecs (5) ─────────────────────────────────────────────────────────────

const specV1 = `openapi: 3.0.3
info: {title: catalog, version: "1.0"}
paths:
  /items:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  sku: {type: string}
                  qty: {type: integer}
  /orders:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required: [sku]
              properties:
                sku: {type: string}
      responses:
        "200": {description: ok}
`

func loadSpec(t *testing.T, content string) *specloader.SpecDoc {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spec.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	doc, err := specloader.NewLoader(0).Load(specloader.SpecSource{SpecPath: path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return &doc
}

func TestDiffSpecs_Happy(t *testing.T) {
	removed := loadSpec(t, strings.ReplaceAll(specV1, "  /orders:\n    post:\n      requestBody:\n        content:\n          application/json:\n            schema:\n              type: object\n              required: [sku]\n              properties:\n                sku: {type: string}\n      responses:\n        \"200\": {description: ok}\n", ""))
	got := DiffSpecs(loadSpec(t, specV1).Info, removed.Info)
	if len(got) == 0 {
		t.Fatalf("removed operation must be breaking, got %+v", got)
	}
	if got[0].Subject != "POST /orders" {
		t.Errorf("subject = %q, want POST /orders", got[0].Subject)
	}
}

func TestDiffSpecs_ResponseFieldRemoved(t *testing.T) {
	v2 := loadSpec(t, strings.ReplaceAll(specV1, "                  qty: {type: integer}\n", ""))
	got := DiffSpecs(loadSpec(t, specV1).Info, v2.Info)
	if len(got) == 0 {
		t.Fatalf("removed response field must be breaking, got %+v", got)
	}
	if got[0].Subject != "GET /items" {
		t.Errorf("subject = %q, want GET /items", got[0].Subject)
	}
}

func TestDiffSpecs_RequestTypeTightened(t *testing.T) {
	v2 := loadSpec(t, strings.ReplaceAll(specV1, "                sku: {type: string}\n      responses:\n        \"200\": {description: ok}\n", "                sku: {type: integer}\n      responses:\n        \"200\": {description: ok}\n"))
	got := DiffSpecs(loadSpec(t, specV1).Info, v2.Info)
	if len(got) == 0 {
		t.Fatalf("request type change string→integer must be breaking, got %+v", got)
	}
}

func TestDiffSpecs_NoChanges(t *testing.T) {
	if got := DiffSpecs(loadSpec(t, specV1).Info, loadSpec(t, specV1).Info); len(got) != 0 {
		t.Errorf("identical specs must yield no breaking changes, got %+v", got)
	}
}

func TestDiffSpecs_AddedOperationIsNotBreaking(t *testing.T) {
	v2 := loadSpec(t, specV1+"  /items/{id}:\n    put:\n      responses:\n        \"204\": {description: ok}\n")
	if got := DiffSpecs(loadSpec(t, specV1).Info, v2.Info); len(got) != 0 {
		t.Errorf("added operation must not be breaking, got %+v", got)
	}
}

// ── AffectedConsumers (3) ─────────────────────────────────────────────────────

func TestAffectedConsumers_Happy(t *testing.T) {
	edges := []store.Edge{
		{Consumer: "checkout-svc", Provider: "catalog", Subject: "POST /orders"},
		{Consumer: "search-svc", Provider: "catalog", Subject: "GET /items"},
	}
	got := AffectedConsumers([]AffectedSubject{{Subject: "POST /orders"}}, edges)
	if len(got) != 1 || got[0] != "checkout-svc" {
		t.Errorf("AffectedConsumers = %v, want [checkout-svc]", got)
	}
}

func TestAffectedConsumers_SubjectNotConsumed(t *testing.T) {
	got := AffectedConsumers([]AffectedSubject{{Subject: "DELETE /all"}}, []store.Edge{
		{Consumer: "search-svc", Provider: "catalog", Subject: "GET /items"},
	})
	if len(got) != 0 {
		t.Errorf("unconsumed subject must affect nobody, got %v", got)
	}
}

func TestAffectedConsumers_MultipleConsumersDedup(t *testing.T) {
	edges := []store.Edge{
		{Consumer: "a-svc", Provider: "p", Subject: "GET /x"},
		{Consumer: "b-svc", Provider: "p", Subject: "GET /x"},
		{Consumer: "a-svc", Provider: "p", Subject: "GET /x"}, // дубль ребра
	}
	got := AffectedConsumers([]AffectedSubject{{Subject: "GET /x"}}, edges)
	if len(got) != 2 {
		t.Errorf("AffectedConsumers = %v, want 2 unique consumers", got)
	}
}
