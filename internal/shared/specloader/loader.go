// Package specloader — I/O-объект загрузки спек (shared; используется срезом
// slice-03-diff-breaking-change). path|URL → OpenAPI 3.x документ (kin-openapi,
// honest reuse — как у pinout-openapi) + sha256-хеш (ключ свежести provenance).
// AsyncAPI-документ детектится и отвергается (MVP E2 — sync только).
package specloader

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/load"
	"gopkg.in/yaml.v3"
)

// Сентинелы (маппинг на HTTP — в адаптере среза 03).
var (
	// ErrSpecNotFound — источник недоступен (файл отсутствует / HTTP не-2xx).
	ErrSpecNotFound = errors.New("SPEC_NOT_FOUND")
	// ErrSpecInvalid — контент не парсится как OpenAPI 3.x.
	ErrSpecInvalid = errors.New("SPEC_INVALID")
	// ErrAsyncImpact — документ является AsyncAPI-спекой (MVP: sync только).
	ErrAsyncImpact = errors.New("ASYNC_IMPACT_NOT_IMPLEMENTED")
)

// SpecSource — ровно один источник спеки: файл ИЛИ URL (oneOf контракта).
type SpecSource struct {
	SpecPath string `json:"spec_path,omitempty"`
	SpecURL  string `json:"spec_url,omitempty"`
}

// NewSpecSource — exactly-one конструктор (valid by construction).
func NewSpecSource(path, url string) (SpecSource, error) {
	if path != "" && url != "" {
		return SpecSource{}, fmt.Errorf("%w: spec_path и spec_url взаимоисключающи", ErrBadRequestSource)
	}
	if path == "" && url == "" {
		return SpecSource{}, fmt.Errorf("%w: нужен ровно один источник (spec_path|spec_url)", ErrBadRequestSource)
	}
	return SpecSource{SpecPath: path, SpecURL: url}, nil
}

// ErrBadRequestSource — источник задан неверно (ветвь BAD_REQUEST среза 03).
var ErrBadRequestSource = errors.New("BAD_REQUEST")

// SpecDoc — загруженная спека: распарсенный документ + sha256 байтов источника.
type SpecDoc struct {
	Info   *load.SpecInfo
	SHA256 string
}

// Loader — автономный I/O-объект (инкапсулирует HTTP-клиент и таймаут).
type Loader struct {
	client  *http.Client
	timeout time.Duration
}

// NewLoader — фабрика; timeout ограничивает загрузку по URL.
func NewLoader(timeout time.Duration) *Loader {
	return &Loader{client: &http.Client{}, timeout: timeout}
}

// Load — загрузить спеку по источнику: bytes → (asyncapi-детект) → kin-openapi.
func (l *Loader) Load(src SpecSource) (SpecDoc, error) {
	data, err := l.fetch(src)
	if err != nil {
		return SpecDoc{}, err
	}
	if isAsyncAPIDoc(data) {
		return SpecDoc{}, fmt.Errorf("%w: AsyncAPI-спека — impact для async-поставщиков в backlog E2", ErrAsyncImpact)
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return SpecDoc{}, fmt.Errorf("%w: %v", ErrSpecInvalid, err)
	}
	sum := sha256.Sum256(data)
	return SpecDoc{
		Info:   &load.SpecInfo{Url: src.displayName(), Spec: doc},
		SHA256: fmt.Sprintf("sha256:%x", sum),
	}, nil
}

// fetch — байты источника: файл (os.ErrNotExist → SPEC_NOT_FOUND) или HTTP GET.
func (l *Loader) fetch(src SpecSource) ([]byte, error) {
	if src.SpecPath != "" {
		data, err := os.ReadFile(src.SpecPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("%w: %s", ErrSpecNotFound, src.SpecPath)
			}
			return nil, fmt.Errorf("%w: %v", ErrSpecInvalid, err)
		}
		return data, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.SpecURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSpecNotFound, err)
	}
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSpecNotFound, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("%w: HTTP %d", ErrSpecNotFound, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSpecNotFound, err)
	}
	return body, nil
}

// isAsyncAPIDoc — детект AsyncAPI по корневому ключу `asyncapi` (YAML или JSON).
func isAsyncAPIDoc(data []byte) bool {
	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return false
	}
	_, ok := root["asyncapi"]
	return ok
}

func (s SpecSource) displayName() string {
	if s.SpecPath != "" {
		return s.SpecPath
	}
	return s.SpecURL
}

// LoaderPort — порт, нужный срезу 03 (структурно удовлетворяется *Loader).
type LoaderPort interface {
	Load(src SpecSource) (SpecDoc, error)
}
