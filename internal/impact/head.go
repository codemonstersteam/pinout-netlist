// head.go — головная ROP-труба среза slice-03-diff-breaking-change
// (contracts.md §ProcessImpact; module-tree.md «Head-pipe pseudocode»).
package impact

import (
	"pinout-netlist/internal/shared/specloader"
	"pinout-netlist/internal/shared/store"
)

// Deps — порты композиционного коруля.
type Deps struct {
	Loader specloader.LoaderPort
	Store  StorePort
}

// StorePort — порт хранилища, нужный срезу.
type StorePort interface {
	ConsumersOf(provider string) []store.Edge
	PrevSpecVersion(provider string) string
	SetLastCapturedHash(provider, hash string) error
}

// ProcessImpact — труба: ParseImpactBody → Load×2 → DiffSpecs → ConsumersOf →
// BreakingChange (+ обновление свежести provenance по хешу новой спеки).
func ProcessImpact(req ImpactRequest, d Deps) (BreakingChange, error) {
	input, err := ParseImpactBody(req.Body)
	if err != nil {
		return BreakingChange{}, err
	}

	from, err := d.Loader.Load(input.FromSpec)
	if err != nil {
		return BreakingChange{}, err
	}
	to, err := d.Loader.Load(input.ToSpec)
	if err != nil {
		return BreakingChange{}, err
	}

	subjects := DiffSpecs(from.Info, to.Info)
	edges := d.Store.ConsumersOf(input.Provider)

	_ = d.Store.SetLastCapturedHash(input.Provider, to.SHA256) // свежесть provenance

	return BreakingChange{
		Provider:          input.Provider,
		FromVersion:       d.Store.PrevSpecVersion(input.Provider),
		ToVersion:         input.ToVersion,
		AffectedConsumers: AffectedConsumers(subjects, edges),
		Details:           detailsOf(subjects),
	}, nil
}
