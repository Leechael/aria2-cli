package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var reqID atomic.Int64

type Config struct {
	Host   string
	Port   int
	Secret string
	Secure bool
}

func LoadConfig() Config {
	c := Config{
		Host: "localhost",
		Port: 6800,
	}
	if v := os.Getenv("ARIA2_RPC_HOST"); v != "" {
		c.Host = v
	}
	if v := os.Getenv("ARIA2_RPC_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Port = p
		}
	}
	if v := os.Getenv("ARIA2_RPC_SECRET"); v != "" {
		c.Secret = v
	}
	if v := os.Getenv("ARIA2_RPC_SECURE"); v == "true" || v == "1" {
		c.Secure = true
	}
	return c
}

func (c Config) Endpoint() string {
	scheme := "http"
	if c.Secure {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d/jsonrpc", scheme, c.Host, c.Port)
}

type Request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type Response struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("aria2 rpc error %d: %s", e.Code, e.Message)
}

type Client struct {
	cfg    Config
	http   *http.Client
	secret string
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		secret: cfg.Secret,
	}
}

func (c *Client) Call(method string, params ...any) (json.RawMessage, error) {
	allParams := make([]any, 0, len(params)+1)
	if c.secret != "" {
		allParams = append(allParams, "token:"+c.secret)
	}
	allParams = append(allParams, params...)

	req := Request{
		JSONRPC: "2.0",
		ID:      strconv.FormatInt(reqID.Add(1), 10),
		Method:  method,
		Params:  allParams,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.http.Post(c.cfg.Endpoint(), "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("rpc call: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var rpcResp Response
	if err := json.Unmarshal(data, &rpcResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}
