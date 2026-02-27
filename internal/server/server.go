package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Config holds the relay server configuration.
type Config struct {
	Port    int
	MaxSize int64         // max blob size in bytes
	MaxTTL  time.Duration // maximum TTL allowed
}

// DefaultConfig returns sensible defaults for the relay server.
func DefaultConfig() Config {
	return Config{
		Port:    3141,
		MaxSize: 10 * 1024 * 1024, // 10MB
		MaxTTL:  time.Hour,
	}
}

// SendRequest is the JSON body for POST /api/send.
type SendRequest struct {
	CodeID string `json:"code_id"`
	Data   string `json:"data"` // base64-encoded encrypted blob
	TTL    int    `json:"ttl"`  // TTL in seconds, 0 = use server default
}

// SendResponse is the JSON response for POST /api/send.
type SendResponse struct {
	OK     bool   `json:"ok"`
	Expiry string `json:"expiry,omitempty"`
	Error  string `json:"error,omitempty"`
}

// ReceiveResponse is the JSON response for GET /api/receive/:id.
type ReceiveResponse struct {
	OK    bool   `json:"ok"`
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// Server is the relay HTTP server.
type Server struct {
	config Config
	store  *Store
	mux    *http.ServeMux
}

// New creates a new relay server.
func New(config Config) *Server {
	s := &Server{
		config: config,
		store:  NewStore(),
		mux:    http.NewServeMux(),
	}
	s.mux.HandleFunc("POST /api/send", s.handleSend)
	s.mux.HandleFunc("GET /api/receive/{id}", s.handleReceive)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	return s
}

// Start starts the relay server and blocks.
func (s *Server) Start() error {
	done := make(chan struct{})
	s.store.StartCleanupLoop(30*time.Second, done)

	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Printf(" git-share relay server listening on %s", addr)
	log.Printf(" Max blob size: %s", formatBytes(s.config.MaxSize))
	log.Printf(" Max TTL: %s", s.config.MaxTTL)

	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	// Enforce size limit
	r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxSize)

	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, SendResponse{Error: "invalid request body"})
		return
	}

	if req.CodeID == "" || req.Data == "" {
		writeJSON(w, http.StatusBadRequest, SendResponse{Error: "code_id and data are required"})
		return
	}

	// Determine TTL
	ttl := s.config.MaxTTL
	if req.TTL > 0 {
		requested := time.Duration(req.TTL) * time.Second
		if requested < ttl {
			ttl = requested
		}
	}

	if !s.store.Put(req.CodeID, []byte(req.Data), ttl) {
		writeJSON(w, http.StatusConflict, SendResponse{Error: "code ID already exists, try again"})
		return
	}

	expiry := time.Now().Add(ttl)
	log.Printf("📦 Stored blob %s (size: %d bytes, TTL: %s)", req.CodeID, len(req.Data), ttl)
	writeJSON(w, http.StatusCreated, SendResponse{OK: true, Expiry: expiry.Format(time.RFC3339)})
}

func (s *Server) handleReceive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ReceiveResponse{Error: "missing code ID"})
		return
	}

	data := s.store.GetAndDelete(id)
	if data == nil {
		writeJSON(w, http.StatusNotFound, ReceiveResponse{Error: "not found or expired"})
		return
	}

	log.Printf("📤 Delivered and deleted blob %s", id)
	writeJSON(w, http.StatusOK, ReceiveResponse{OK: true, Data: string(data)})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"blobs": s.store.Count(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
