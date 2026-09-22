package store

import (
	"os"
	"path/filepath"
	"testing"
)

// TestApplyAndGraph_Happy — 1 happy: apply двух рёбер → граф с последним вердиктом.
func TestApplyAndGraph_Happy(t *testing.T) {
	s := New("")
	e1 := Edge{Consumer: "search-svc", Provider: "catalog", Subject: "GET /items", Interaction: "sync"}
	e2 := Edge{Consumer: "checkout-svc", Provider: "catalog", Subject: "POST /orders", Interaction: "sync"}
	u := Update{Edges: []Edge{e1, e2}, CapturedHash: "sha256:1"}
	u.SpecVersion.Provider = "catalog"
	u.SpecVersion.Version = "1.0.0"
	u.Verdicts = []VerdictRecord{
		{Edge: e1, Compatible: true, ProviderVersion: "1.0.0", CapturedHash: "sha256:1", At: "2026-09-22T10:00:00Z"},
		{Edge: e2, Compatible: false, Errors: []ReportError{{Code: "R", Subject: "POST /orders"}}, ProviderVersion: "1.0.0", CapturedHash: "sha256:1", At: "2026-09-22T10:00:00Z"},
	}
	if _, err := s.Apply(u); err != nil {
		t.Fatalf("Apply() error: %v", err)
	}
	g := s.Graph()
	if len(g) != 2 {
		t.Fatalf("Graph() len = %d, want 2", len(g))
	}
	bySubject := map[string]bool{}
	for _, e := range g {
		bySubject[e.Subject] = e.Compatible
	}
	if !bySubject["GET /items"] || bySubject["POST /orders"] {
		t.Errorf("verdict projection wrong: %+v", g)
	}
	if s.ConsumersOf("catalog") == nil || len(s.ConsumersOf("catalog")) != 2 {
		t.Errorf("ConsumersOf = %v", s.ConsumersOf("catalog"))
	}
	if s.PrevSpecVersion("catalog") != "1.0.0" || s.LastCapturedHash("catalog") != "sha256:1" {
		t.Errorf("spec version/hash not stored")
	}
}

// TestGraph_Stale — ветвь: хеш вердикта ≠ последний хеш поставщика → stale.
func TestGraph_Stale(t *testing.T) {
	s := New("")
	e := Edge{Consumer: "c", Provider: "p", Subject: "GET /x", Interaction: "sync"}
	u := Update{Edges: []Edge{e}, CapturedHash: "sha256:old"}
	u.SpecVersion.Provider = "p"
	u.Verdicts = []VerdictRecord{{Edge: e, Compatible: true, CapturedHash: "sha256:old", At: "2026-09-22T10:00:00Z"}}
	_, _ = s.Apply(u)
	if err := s.SetLastCapturedHash("p", "sha256:new"); err != nil {
		t.Fatalf("SetLastCapturedHash: %v", err)
	}
	g := s.Graph()
	if !g[0].Stale {
		t.Errorf("edge must be stale after provider hash changed: %+v", g[0])
	}
}

// TestSnapshotRoundtrip — ветвь: снапшот переживает рестарт.
func TestSnapshotRoundtrip(t *testing.T) {
	file := filepath.Join(t.TempDir(), "netlist.json")
	s1 := New(file)
	e := Edge{Consumer: "c", Provider: "p", Subject: "GET /x", Interaction: "sync"}
	u := Update{Edges: []Edge{e}}
	u.SpecVersion.Provider = "p"
	u.Verdicts = []VerdictRecord{{Edge: e, Compatible: true, At: "2026-09-22T10:00:00Z"}}
	if _, err := s1.Apply(u); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("snapshot not written: %v", err)
	}
	s2 := New(file)
	if g := s2.Graph(); len(g) != 1 || g[0].Subject != "GET /x" {
		t.Errorf("snapshot roundtrip lost edge: %+v", g)
	}
}

// TestApply_Idempotent — ветвь: повторный apply того же ребра не растит граф.
func TestApply_Idempotent(t *testing.T) {
	s := New("")
	e := Edge{Consumer: "c", Provider: "p", Subject: "GET /x", Interaction: "sync"}
	u := Update{Edges: []Edge{e}}
	n1, _ := s.Apply(u)
	n2, _ := s.Apply(u)
	if n1 != 1 || n2 != 0 {
		t.Errorf("created = %d then %d, want 1 then 0", n1, n2)
	}
	if len(s.Graph()) != 1 {
		t.Errorf("graph grew on idempotent apply")
	}
}
