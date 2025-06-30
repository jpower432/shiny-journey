package claims

import "sync"

type Store struct {
	mu     sync.RWMutex
	claims map[string]ConformanceClaim
}

func NewStore() *Store {
	return &Store{
		claims: make(map[string]ConformanceClaim),
	}
}

func (s *Store) Add(claim ConformanceClaim) {
	s.mu.Lock()
	s.claims[claim.ClaimID] = claim
	s.mu.Unlock()
}

func (s *Store) GetClaims() []ConformanceClaim {
	claims := make([]ConformanceClaim, 0, len(s.claims))
	for _, claim := range s.claims {
		claims = append(claims, claim)
	}
	return claims
}
