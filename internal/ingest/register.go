// register.go — wiring среза slice-01-ingest-report: конкретные реализации портов.
package ingest

import "pinout-netlist/internal/shared/store"

// NewDeps — сборка портов среза вокруг хранилища.
func NewDeps(s *store.Store) Deps {
	return Deps{Store: s}
}
