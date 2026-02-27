package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for the git-share relay server.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// SendRequest matches the server's expected JSON body.
type SendRequest struct {
	CodeID string `json:"code_id"`
	Data   string `json:"data"`
	TTL    int    `json:"ttl"`
}

// SendResponse matches the server's JSON response.
type SendResponse struct {
	OK     bool   `json:"ok"`
	Expiry string `json:"expiry,omitempty"`
	Error  string `json:"error,omitempty"`
}

// ReceiveResponse matches the server's JSON response.
type ReceiveResponse struct {
	OK    bool   `json:"ok"`
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// New creates a new relay client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Send uploads an encrypted blob to the relay server.
func (c *Client) Send(codeID string, data string, ttlSeconds int) (*SendResponse, error) {
	reqBody := SendRequest{
		CodeID: codeID,
		Data:   data,
		TTL:    ttlSeconds,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/send", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("connecting to relay server at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var sendResp SendResponse
	if err := json.Unmarshal(respBody, &sendResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if !sendResp.OK {
		return nil, fmt.Errorf("server error: %s", sendResp.Error)
	}

	return &sendResp, nil
}

// Receive downloads and consumes an encrypted blob from the relay server.
func (c *Client) Receive(codeID string) (string, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/receive/" + codeID)
	if err != nil {
		return "", fmt.Errorf("connecting to relay server at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var recvResp ReceiveResponse
	if err := json.Unmarshal(respBody, &recvResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if !recvResp.OK {
		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("patch not found — it may have already been received or expired")
		}
		return "", fmt.Errorf("server error: %s", recvResp.Error)
	}

	return recvResp.Data, nil
}
