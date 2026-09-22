// register.go — wiring среза slice-03-diff-breaking-change.
package impact

import (
	"time"

	"pinout-netlist/internal/shared/specloader"
)

// NewDeps — сборка портов среза (загрузчик спек + хранилище).
func NewDeps(loader *specloader.Loader, st StorePort) Deps {
	return Deps{Loader: loader, Store: st}
}

// DefaultLoaderTimeout — таймаут загрузки спеки по URL (15s; конфигурируется
// при сборке Deps в cmd/app при необходимости).
const DefaultLoaderTimeout = 15 * time.Second
