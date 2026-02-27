package server

import (
	"sync"
	"time"
)

// Blob represents an encrypted patch stored on the relay server.
type Blob struct {
	Data      []byte
	CreatedAt time.Time
	TTL       time.Duration
}

// Store is a thread-safe in-memory blob store with TTL and one-time-use semantics.
type Store struct {
	mu    sync.RWMutex
	blobs map[string]*Blob
}

// NewStore creates a new empty blob store.
func NewStore() *Store {
	return &Store{
		blobs: make(map[string]*Blob),
	}
}

// Put stores an encrypted blob with the given TTL.
// Returns false if the code ID already exists.
func (s *Store) Put(codeID string, data []byte, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.blobs[codeID]; exists {
		return false
	}

	s.blobs[codeID] = &Blob{
		Data:      data,
		CreatedAt: time.Now(),
		TTL:       ttl,
	}
	return true
}

// GetAndDelete atomically retrieves and deletes a blob (one-time use).
// Returns nil if the blob doesn't exist or has expired.
func (s *Store) GetAndDelete(codeID string) []byte {
	s.mu.Lock()
	defer s.mu.Unlock()

	blob, exists := s.blobs[codeID]
	if !exists {
		return nil
	}

	// Check TTL
	if time.Since(blob.CreatedAt) > blob.TTL {
		delete(s.blobs, codeID)
		return nil
	}

	data := blob.Data
	delete(s.blobs, codeID)
	return data
}

// Cleanup removes all expired blobs. Should be called periodically.
func (s *Store) Cleanup() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	now := time.Now()
	for id, blob := range s.blobs {
		if now.Sub(blob.CreatedAt) > blob.TTL {
			delete(s.blobs, id)
			removed++
		}
	}
	return removed
}

// Count returns the number of currently stored blobs.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.blobs)
}

// StartCleanupLoop starts a background goroutine that periodically cleans up expired blobs.
func (s *Store) StartCleanupLoop(interval time.Duration, done <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.Cleanup()
			case <-done:
				return
			}
		}
	}()
}
