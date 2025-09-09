package session

import (
	"sync"
	"time"
)

// InMemorySessionStore manages revoked JWT tokens in memory
type InMemorySessionStore struct {
	// For simplicity, we store token string in a set to invalidate on logout.
	// In production, use a blacklist with TTL in Redis, or rotate tokens with server-side sessions.
	revoked map[string]time.Time
	mutex   sync.RWMutex
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		revoked: make(map[string]time.Time),
	}
}

// RevokeToken adds a token to the revoked list
func (s *InMemorySessionStore) RevokeToken(token string, expiration time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.revoked[token] = expiration
}

// IsTokenRevoked checks if a token is in the revoked list
func (s *InMemorySessionStore) IsTokenRevoked(token string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	expiration, exists := s.revoked[token]
	if !exists {
		return false
	}
	
	// Clean up expired tokens
	if time.Now().After(expiration) {
		delete(s.revoked, token)
		return false
	}
	
	return true
}

// CleanExpiredTokens removes expired tokens from the revoked list
func (s *InMemorySessionStore) CleanExpiredTokens() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	now := time.Now()
	for token, expiration := range s.revoked {
		if now.After(expiration) {
			delete(s.revoked, token)
		}
	}
}
