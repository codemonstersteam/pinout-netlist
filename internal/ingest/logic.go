// logic.go — чистая логика среза: вывод субъектов и совместимости ребра.
// contracts.md §SubjectsOf; messages.md «Следствия варианта A» (агрегатный compatible).
package ingest

// SubjectsOf — дедуплицированное объединение: subjects (scope вызывающего) ∪
// errors[].subject (субъекты нарушений) ∪ uncovered_* (поверхность вне scope).
// Порядок устойчив: сначала subjects, затем ошибки, затем uncovered — как даны.
func SubjectsOf(r Report, subjects []string) []string {
	seen := make(map[string]struct{}, len(subjects)+len(r.errors)+len(r.uncovered))
	out := make([]string, 0, len(subjects))
	add := func(s string) {
		if s == "" {
			return
		}
		if _, dup := seen[s]; dup {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	for _, s := range subjects {
		add(s)
	}
	for _, e := range r.errors {
		add(e.Subject)
	}
	for _, u := range r.uncovered {
		add(u)
	}
	return out
}

// EdgeCompatible — совместимость ребра = «в отчёте нет ошибки с этим субъектом»
// (агрегатный compatible отчёта разложен по субъектам; ловушка «все рёбра красные»
// закрыта by construction). Отчёт без ошибок ⇒ все рёбра зелёные.
func EdgeCompatible(r Report, subject string) bool {
	for _, e := range r.errors {
		if e.Subject == subject {
			return false
		}
	}
	return true
}

// ErrorsOfSubject — ошибки отчёта с данным субъектом (вердикт ребра).
func ErrorsOfSubject(r Report, subject string) []RawError {
	var out []RawError
	for _, e := range r.errors {
		if e.Subject == subject {
			out = append(out, e)
		}
	}
	return out
}
