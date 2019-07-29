package main

type Results struct {
	files []string
	Paths map[string]bool
	URLs  map[string]bool
}

func NewResults() *Results {
	return &Results{
		Paths: map[string]bool{},
		URLs:  map[string]bool{},
	}
}

func (r *Results) Add(t string, e string) *Results {
	switch t {
	case "filename":
		r.Paths[e] = true
	case "url":
		r.URLs[e] = true
	}
	return r
}

func (r *Results) Merge(s *Results) *Results {
	for p, _ := range s.Paths {
		r.Paths[p] = true
	}
	for u, _ := range s.URLs {
		r.URLs[u] = true
	}
	return r
}
