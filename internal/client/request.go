package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	serverScheme = "http://" // just for now (dev) - need to figure out letsEncrypt certs
)

// Signed, with body hash
func ufoSignedRequest[T any](c *Client, method, url string, reqBody any, headers map[string]string) (T, error) {
	req, err := buildJSONRequest(method, url, reqBody, headers)
	if err != nil {
		var zero T
		return zero, err
	}

	if err := c.Sign(req, time.Now().Unix(), reqBody != nil); err != nil {
		var zero T
		return zero, err
	}

	return sendAndDecode[T](c, req)
}

// Unsigned
func ufoPublicRequest[T any](c *Client, method, url string, reqBody any, headers map[string]string) (T, error) {
	req, err := buildJSONRequest(method, url, reqBody, headers)
	if err != nil {
		var zero T
		return zero, err
	}

	return sendAndDecode[T](c, req)
}

// Signed, no nody hash
func ufoStreamRequest(c *Client, method, url string, body io.Reader, size int64, headers map[string]string) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("Error building request: %w", err)
	}

	req.ContentLength = size
	req.Header.Set("Content-Type", "application/octet-stream")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if err := c.Sign(req, time.Now().Unix(), false); err != nil {
		return err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("Request returned status (%d)", res.StatusCode)
	}
	return nil
}

// Shared Helpers

func buildJSONRequest(method, url string, body any, headers map[string]string) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = serverScheme + url
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("Error building request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func sendAndDecode[T any](c *Client, req *http.Request) (T, error) {
	var resData T
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return resData, fmt.Errorf("Error sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return resData, fmt.Errorf("Server error (%d)", res.StatusCode)
	}

	// Handle empty response bodies (like 204 No Content)
	if res.StatusCode == http.StatusNoContent {
		return resData, nil
	}

	if err = json.NewDecoder(res.Body).Decode(&resData); err != nil {
		return resData, fmt.Errorf("Error decoding response: %w", err)
	}
	return resData, nil
}
