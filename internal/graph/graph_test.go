package graph

import (
	"testing"

	"pinout-netlist/internal/shared/store"
)

// fakeGraphPort — порт с готовыми рёбрами (юнит проекции, не хранилища).
type fakeGraphPort struct{ edges []store.GraphEdge }

func (f fakeGraphPort) Graph() []store.GraphEdge { return f.edges }

// TestProcessGraph_Happy — 1 happy: рёбра проецируются с последним вердиктом.
func TestProcessGraph_Happy(t *testing.T) {
	in := []store.GraphEdge{{
		Edge:            store.Edge{Consumer: "search-svc", Provider: "catalog", Subject: "GET /items", Interaction: "sync"},
		Compatible:      true,
		LastVerdictAt:   "2026-09-22T10:00:00Z",
		ProviderVersion: "1.0.0",
	}}
	got := ProcessGraph(Deps{Store: fakeGraphPort{edges: in}})
	if len(got.Edges) != 1 || got.Edges[0].Consumer != "search-svc" || !got.Edges[0].Compatible {
		t.Errorf("projection = %+v", got.Edges)
	}
}

// TestProcessGraph_EmptyIsNotNull — ветвь: пустой граф → edges: [] (не null).
func TestProcessGraph_EmptyIsNotNull(t *testing.T) {
	got := ProcessGraph(Deps{Store: fakeGraphPort{}})
	if got.Edges == nil || len(got.Edges) != 0 {
		t.Errorf("empty graph must be [], got %#v", got.Edges)
	}
}
