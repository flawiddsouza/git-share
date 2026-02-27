package server

import (
	"testing"
	"time"
)

func TestStorePutAndGet(t *testing.T) {
	s := NewStore()
	data := []byte("encrypted-blob")

	ok := s.Put("abc123", data, time.Hour)
	if !ok {
		t.Fatal("Put should succeed")
	}

	got := s.GetAndDelete("abc123")
	if got == nil {
		t.Fatal("GetAndDelete should return data")
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestStoreOneTimeUse(t *testing.T) {
	s := NewStore()
	s.Put("abc123", []byte("data"), time.Hour)

	// First get should succeed
	got := s.GetAndDelete("abc123")
	if got == nil {
		t.Fatal("first GetAndDelete should return data")
	}

	// Second get should return nil
	got = s.GetAndDelete("abc123")
	if got != nil {
		t.Error("second GetAndDelete should return nil (one-time use)")
	}
}

func TestStoreTTLExpiry(t *testing.T) {
	s := NewStore()
	s.Put("abc123", []byte("data"), 1*time.Millisecond)

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	got := s.GetAndDelete("abc123")
	if got != nil {
		t.Error("GetAndDelete should return nil after TTL expiry")
	}
}

func TestStoreDuplicateCodeID(t *testing.T) {
	s := NewStore()
	s.Put("abc123", []byte("data1"), time.Hour)

	ok := s.Put("abc123", []byte("data2"), time.Hour)
	if ok {
		t.Error("duplicate Put should return false")
	}
}

func TestStoreCleanup(t *testing.T) {
	s := NewStore()
	s.Put("expired", []byte("data"), 1*time.Millisecond)
	s.Put("fresh", []byte("data"), time.Hour)

	time.Sleep(10 * time.Millisecond)
	removed := s.Cleanup()

	if removed != 1 {
		t.Errorf("cleanup should remove 1 blob, removed %d", removed)
	}
	if s.Count() != 1 {
		t.Errorf("should have 1 blob remaining, got %d", s.Count())
	}
}

func TestStoreNotFound(t *testing.T) {
	s := NewStore()
	got := s.GetAndDelete("nonexistent")
	if got != nil {
		t.Error("GetAndDelete for nonexistent key should return nil")
	}
}
