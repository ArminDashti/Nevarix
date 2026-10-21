package history

import (
	"sync"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

type Store struct {
	mu      sync.Mutex
	max     int
	series  map[string][]sysreport.Sample
}

func New(max int) *Store {
	if max < 10 {
		max = 10
	}
	return &Store{max: max, series: map[string][]sysreport.Sample{}}
}

func (s *Store) Push(key string, sample sysreport.Sample) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hist := append(s.series[key], sample)
	if len(hist) > s.max {
		hist = hist[len(hist)-s.max:]
	}
	s.series[key] = hist
}

func (s *Store) Get(key string) []sysreport.Sample {
	s.mu.Lock()
	defer s.mu.Unlock()
	src := s.series[key]
	out := make([]sysreport.Sample, len(src))
	copy(out, src)
	return out
}
