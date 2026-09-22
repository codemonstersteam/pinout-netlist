// head.go — головная ROP-труба среза slice-01-ingest-report (contracts.md
// §ProcessIngest; module-tree.md «Head-pipe pseudocode»). Линейная, без
// собственной логики и ветвлений: шаг с сентинелом короткозит и поднимает
// ошибку нетрансформированной; маппинг сентинел → HTTP — только в adapter.go.
package ingest

import (
	"pinout-netlist/internal/shared/store"
)

// IngestRequest — единственный вход среза (тело POST /reports).
type IngestRequest struct {
	Body []byte
}

// IngestAccepted — единственный результат (202).
type IngestAccepted struct {
	EdgesAccepted int `json:"edges_accepted"`
}

// StorePort — порт хранилища, нужный срезу (structural: *store.Store удовлетворяет).
type StorePort interface {
	Apply(u store.Update) (int, error)
}

// Deps — порты композиционного корня (register.go поставляет реализации).
type Deps struct {
	Store StorePort
}

// ProcessIngest — труба: ParseIngestBody → SubjectsOf → построение рёбер/вердиктов →
// Store.Apply. Совместимость ребра = «нет ошибки с этим субъектом» (EdgeCompatible).
func ProcessIngest(req IngestRequest, d Deps) (IngestAccepted, error) {
	input, err := ParseIngestBody(req.Body)
	if err != nil {
		return IngestAccepted{}, err
	}

	subjects := SubjectsOf(input.Report, input.Subjects)

	var edges []store.Edge
	var verdicts []store.VerdictRecord
	for _, subject := range subjects {
		edge, err := NewEdge(input.Report, subject)
		if err != nil {
			return IngestAccepted{}, err
		}
		verdict, err := NewVerdictRecord(input.Report, subject)
		if err != nil {
			return IngestAccepted{}, err
		}
		edges = append(edges, edge)
		verdicts = append(verdicts, verdict)
	}

	u := store.Update{Edges: edges, Verdicts: verdicts, CapturedHash: input.Report.capturedHash}
	u.SpecVersion.Provider = input.Report.provider
	u.SpecVersion.Version = input.Report.providerVersion

	created, err := d.Store.Apply(u)
	if err != nil {
		return IngestAccepted{}, err
	}
	return IngestAccepted{EdgesAccepted: created}, nil
}
