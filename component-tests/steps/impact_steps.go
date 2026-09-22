// impact_steps — срезовые godog-шаги slice-03-diff-breaking-change:
// POST /impact с телом из файла-фикстуры.
package steps

import (
	"net/http"
	"os"

	"github.com/cucumber/godog"
)

func (w *World) registerImpactSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^клиент отправляет POST /impact с телом из файла (\S+)$`, w.postImpactFromFile)
}

// postImpactFromFile — POST /impact с телом из bind-mounted файла фикстуры.
func (w *World) postImpactFromFile(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return w.doRequest(http.MethodPost, "/impact", body)
}
