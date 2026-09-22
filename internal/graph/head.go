// Package graph — слайс slice-02-query-graph: проекция хранилища в текущий граф
// пар consumer↔provider с последним вердиктом каждого ребра.
// contracts.md §QueryGraph; module-tree.md (один вход — один результат, ветвей
// отказа нет: пустой граф — валидный ответ).
package graph

import "pinout-netlist/internal/shared/store"

// GraphResponse — ответ GET /graph (openapi.yaml Graph).
type GraphResponse struct {
	Edges []store.GraphEdge `json:"edges"`
}

// GraphPort — порт хранилища, нужный срезу.
type GraphPort interface {
	Graph() []store.GraphEdge
}

// Deps — порт композиционного коруля.
type Deps struct {
	Store GraphPort
}

// ProcessGraph — единственный шаг: чтение + проекция.
func ProcessGraph(d Deps) GraphResponse {
	return GraphResponse{Edges: project(d.Store.Graph())}
}

// project — проекция записей хранилища в ответ; пустое хранилище → edges: []
// (не null — required-поле контракта), сортировка уже в Store.
func project(edges []store.GraphEdge) []store.GraphEdge {
	if edges == nil {
		return []store.GraphEdge{}
	}
	return edges
}
