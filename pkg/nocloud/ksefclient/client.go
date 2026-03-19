package ksefclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/slntopp/nocloud-proto/billing"
	"google.golang.org/protobuf/encoding/protojson"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type ValidationResponse struct {
	Valid  bool              `json:"valid"`
	Errors []json.RawMessage `json:"errors,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (c *Client) ValidateInvoice(
	ctx context.Context,
	token string,
	invoice *billing.Invoice,
) (*ValidationResponse, int, error) {
	if invoice == nil {
		return nil, 0, fmt.Errorf("invoice is nil")
	}
	if strings.TrimSpace(token) == "" {
		return nil, 0, fmt.Errorf("token is empty")
	}

	body, err := protojson.Marshal(invoice)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal invoice: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/invoice/validate",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusUnprocessableEntity:
		var out ValidationResponse
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, resp.StatusCode, fmt.Errorf(
				"decode validation response: %w; body=%s",
				err, string(respBody),
			)
		}
		return &out, resp.StatusCode, nil

	default:
		var apiErr ErrorResponse
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error != "" {
			return nil, resp.StatusCode, fmt.Errorf("api error: %s", apiErr.Error)
		}

		return nil, resp.StatusCode, fmt.Errorf(
			"unexpected status %d: %s",
			resp.StatusCode,
			string(respBody),
		)
	}
}
