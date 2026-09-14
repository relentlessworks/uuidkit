package store

import (
	"sync/atomic"
)

// Stats tracks in-memory usage counters.
type Stats struct {
	UUIDGenerated int64
	ULIDGenerated int64
	Parsed        int64
	Validated     int64
}

// Store is an in-memory usage tracker. No persistence needed for a stateless service.
type Store struct {
	uuidGen  int64
	ulidGen  int64
	parsed   int64
	validated int64
}

// New creates a new in-memory store.
func New() *Store {
	return &Store{}
}

func (s *Store) IncrUUID()  { atomic.AddInt64(&s.uuidGen, 1) }
func (s *Store) IncrULID()  { atomic.AddInt64(&s.ulidGen, 1) }
func (s *Store) IncrParse() { atomic.AddInt64(&s.parsed, 1) }
func (s *Store) IncrValid() { atomic.AddInt64(&s.validated, 1) }

// GetStats returns a snapshot of current usage counters.
func (s *Store) GetStats() Stats {
	return Stats{
		UUIDGenerated: atomic.LoadInt64(&s.uuidGen),
		ULIDGenerated: atomic.LoadInt64(&s.ulidGen),
		Parsed:        atomic.LoadInt64(&s.parsed),
		Validated:     atomic.LoadInt64(&s.validated),
	}
}
