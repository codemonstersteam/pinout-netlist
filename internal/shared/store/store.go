// Package store — I/O-объект хранилища pinout-netlist (shared: используют срезы
// 01/02/03). In-memory с JSON-снапшотом (CGO_ENABLED=0, без внешней БД); движок
// за интерфейсом — замена не задевает логику срезов (docs/design/messages.md).
// Port: UpsertEdge / AppendVerdict / AppendSpecVersion / Graph / ConsumersOf /
// PrevSpecVersion / LastCapturedHash / SetLastCapturedHash.
package store

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

// ReportError — ошибка вердикта (эхо errors[] канона 1.1: code/message/subject).
type ReportError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Subject string `json:"subject"`
}

// Edge — ребро графа consumer→provider по субъекту (операция/канал).
type Edge struct {
	Consumer    string `json:"consumer"`
	Provider    string `json:"provider"`
	Subject     string `json:"subject"`
	Interaction string `json:"interaction"`
}

// VerdictRecord — факт проверки ребра во времени (append-only история).
type VerdictRecord struct {
	Edge            Edge          `json:"edge"`
	Compatible      bool          `json:"compatible"`
	Errors          []ReportError `json:"errors"`
	ProviderVersion string        `json:"provider_version"`
	CapturedHash    string        `json:"captured_hash"`
	At              string        `json:"at"` // generated_at отчёта
}

// GraphEdge — ребро с последним вердиктом (ответ GET /graph).
type GraphEdge struct {
	Edge
	Compatible      bool   `json:"compatible"`
	LastVerdictAt   string `json:"last_verdict_at"`
	ProviderVersion string `json:"provider_version"`
	Stale           bool   `json:"stale,omitempty"`
}

// Update — атомарное обновление хранилища от одного отчёта (ingest).
type Update struct {
	Edges       []Edge          // upsert (идемпотентно по ключу ребра)
	Verdicts    []VerdictRecord // append по каждому ребру
	SpecVersion struct {
		Provider string
		Version  string
	}
	CapturedHash string // свежесть provenance поставщика
}

// Store — потокобезопасное in-memory хранилище с опциональным JSON-снапшотом.
type Store struct {
	mu           sync.Mutex
	edges        map[Edge]struct{}
	verdicts     []VerdictRecord // append-only
	lastVerdict  map[Edge]VerdictRecord
	specVersions map[string]string // provider → последняя версия
	lastHash     map[string]string // provider → последний captured_hash
	file         string            // путь снапшота ("" = только память)
}

// New — конструктор хранилища; file ≠ "" → снапшот переживает рестарт.
func New(file string) *Store {
	s := &Store{
		edges:        map[Edge]struct{}{},
		lastVerdict:  map[Edge]VerdictRecord{},
		specVersions: map[string]string{},
		lastHash:     map[string]string{},
		file:         file,
	}
	if file != "" {
		if data, err := os.ReadFile(file); err == nil {
			var snap struct {
				Edges        []Edge            `json:"edges"`
				Verdicts     []VerdictRecord   `json:"verdicts"`
				SpecVersions map[string]string `json:"spec_versions"`
				LastHash     map[string]string `json:"last_hash"`
			}
			if json.Unmarshal(data, &snap) == nil {
				for _, e := range snap.Edges {
					s.edges[e] = struct{}{}
				}
				s.verdicts = snap.Verdicts
				for _, v := range snap.Verdicts {
					if cur, ok := s.lastVerdict[v.Edge]; !ok || v.At >= cur.At {
						s.lastVerdict[v.Edge] = v
					}
				}
				if snap.SpecVersions != nil {
					s.specVersions = snap.SpecVersions
				}
				if snap.LastHash != nil {
					s.lastHash = snap.LastHash
				}
			}
		}
	}
	return s
}

// Apply — применить обновление одного отчёта: идемпотентный upsert рёбер +
// append вердиктов + обновление версии/хеша поставщика. Возвращает число
// созданных/обновлённых рёбер.
func (s *Store) Apply(u Update) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	created := 0
	for _, e := range u.Edges {
		if _, ok := s.edges[e]; !ok {
			created++
		}
		s.edges[e] = struct{}{}
	}
	s.verdicts = append(s.verdicts, u.Verdicts...)
	for _, v := range u.Verdicts {
		if cur, ok := s.lastVerdict[v.Edge]; !ok || v.At >= cur.At {
			s.lastVerdict[v.Edge] = v
		}
	}
	if u.SpecVersion.Provider != "" {
		s.specVersions[u.SpecVersion.Provider] = u.SpecVersion.Version
	}
	if u.CapturedHash != "" && u.SpecVersion.Provider != "" {
		s.lastHash[u.SpecVersion.Provider] = u.CapturedHash
	}
	return created, s.snapshot()
}

// Graph — все рёбра с последним вердиктом; stale = хеш вердикта ≠ последний
// известный хеш спеки поставщика (свежесть provenance, messages.md §3).
func (s *Store) Graph() []GraphEdge {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]GraphEdge, 0, len(s.edges))
	for e := range s.edges {
		ge := GraphEdge{Edge: e}
		if v, ok := s.lastVerdict[e]; ok {
			ge.Compatible = v.Compatible
			ge.LastVerdictAt = v.At
			ge.ProviderVersion = v.ProviderVersion
			ge.Stale = v.CapturedHash != s.lastHash[e.Provider]
		}
		out = append(out, ge)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Consumer != out[j].Consumer {
			return out[i].Consumer < out[j].Consumer
		}
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return out[i].Subject < out[j].Subject
	})
	return out
}

// ConsumersOf — рёбра потребителей данного поставщика.
func (s *Store) ConsumersOf(provider string) []Edge {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Edge
	for e := range s.edges {
		if e.Provider == provider {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Consumer != out[j].Consumer {
			return out[i].Consumer < out[j].Consumer
		}
		return out[i].Subject < out[j].Subject
	})
	return out
}

// PrevSpecVersion — последняя известная версия спеки поставщика ("" если нет).
func (s *Store) PrevSpecVersion(provider string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.specVersions[provider]
}

// LastCapturedHash / SetLastCapturedHash — ключ свежести provenance.
func (s *Store) LastCapturedHash(provider string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastHash[provider]
}

func (s *Store) SetLastCapturedHash(provider, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if provider == "" || hash == "" {
		return nil
	}
	s.lastHash[provider] = hash
	return s.snapshot()
}

// snapshot — атомарная запись JSON-снапшота (временный файл + rename).
func (s *Store) snapshot() error {
	if s.file == "" {
		return nil
	}
	snap := struct {
		Edges        []Edge            `json:"edges"`
		Verdicts     []VerdictRecord   `json:"verdicts"`
		SpecVersions map[string]string `json:"spec_versions"`
		LastHash     map[string]string `json:"last_hash"`
	}{}
	for e := range s.edges {
		snap.Edges = append(snap.Edges, e)
	}
	sort.Slice(snap.Edges, func(i, j int) bool {
		if snap.Edges[i].Consumer != snap.Edges[j].Consumer {
			return snap.Edges[i].Consumer < snap.Edges[j].Consumer
		}
		return snap.Edges[i].Subject < snap.Edges[j].Subject
	})
	snap.Verdicts = s.verdicts
	snap.SpecVersions = s.specVersions
	snap.LastHash = s.lastHash
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.file + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.file)
}
