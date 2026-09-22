// adapter.go — ingress-адаптер среза 03: JSON-декод запроса + маппинг
// сентинел → HTTP-статус (единственное место маппинга; contracts.md «Error model»).
package impact

import (
	"encoding/json"
	"errors"
	"net/http"

	"pinout-netlist/internal/shared/specloader"
)

// ParseImpactBody — декод тела → ImpactInput (источники exactly-one).
func ParseImpactBody(body []byte) (ImpactInput, error) {
	var env impactEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return ImpactInput{}, ErrBadRequest
	}
	if env.Provider == "" || env.ToVersion == "" {
		return ImpactInput{}, ErrBadRequest
	}
	from, err := specloader.NewSpecSource(env.FromSpec.SpecPath, env.FromSpec.SpecURL)
	if err != nil {
		return ImpactInput{}, ErrBadRequest
	}
	to, err := specloader.NewSpecSource(env.ToSpec.SpecPath, env.ToSpec.SpecURL)
	if err != nil {
		return ImpactInput{}, ErrBadRequest
	}
	return ImpactInput{
		Provider:  env.Provider,
		ToVersion: env.ToVersion,
		FromSpec:  from,
		ToSpec:    to,
	}, nil
}

// HTTPStatus — маппинг сентинелов среза на статусы (adapter-таблица failure-map).
func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest // 400
	case errors.Is(err, specloader.ErrSpecNotFound):
		return http.StatusNotFound // 404
	case errors.Is(err, specloader.ErrSpecInvalid):
		return http.StatusUnprocessableEntity // 422
	case errors.Is(err, specloader.ErrAsyncImpact):
		return http.StatusNotImplemented // 501
	default:
		return http.StatusInternalServerError // 500
	}
}

// ErrorCode — чистый код сентинела для error.code ответа (детали — в message).
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrBadRequest):
		return "BAD_REQUEST"
	case errors.Is(err, specloader.ErrSpecNotFound):
		return "SPEC_NOT_FOUND"
	case errors.Is(err, specloader.ErrSpecInvalid):
		return "SPEC_INVALID"
	case errors.Is(err, specloader.ErrAsyncImpact):
		return "ASYNC_IMPACT_NOT_IMPLEMENTED"
	default:
		return "INTERNAL"
	}
}
