// Package kbhttp provides a simple HTTP client for testing with Katabole.
//
// For a canonical usage example see https://github.com/katabole/kbexample/tree/main/actions
//
// Example:
//
//	client := kbhttp.NewClient(kbhttp.ClientConfig{BaseURL: "http://localhost:3000/"})
//
//	var user models.User
//	if err := f.Client.GetJSON("/users/12345", &user); err != nil {
//		t.Fatalf("failed to get user: %v", err)
//	}
//
//	user.Name = "Tom"
//	if err := f.Client.PutJSON("/users/12345", user, nil); err != nil {
//		t.Fatalf("failed to update user: %v", err)
//	}
package kbhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"unicode/utf8"
)

type ClientConfig struct {
	BaseURL *url.URL
}

type Client struct {
	*http.Client
	config ClientConfig
}

func NewClient(config ClientConfig) *Client {
	return &Client{
		Client: http.DefaultClient,
		config: config,
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.config.BaseURL == nil {
		return c.Client.Do(req)
	}

	// Clone so we don't modify the argument, just in case
	r := req.Clone(req.Context())
	r.URL.Scheme = c.config.BaseURL.Scheme
	r.URL.Host = c.config.BaseURL.Host
	r.URL.Path = path.Join(c.config.BaseURL.Path, r.URL.Path)
	r.URL.RawPath = path.Join(c.config.BaseURL.RawPath, r.URL.RawPath)
	return c.Client.Do(r)
}

// JSON
//

func (c *Client) DoJSON(req *http.Request, target any) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("got %d code and failed to read response body: %w", resp.StatusCode, err)
		}

		if !utf8.Valid(body) {
			return fmt.Errorf("got %d code and %d bytes of binary data", resp.StatusCode, len(body))
		}
		return fmt.Errorf("got %d code and response: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&target); err != nil {
		return fmt.Errorf("got %d code and failed to decode response body: %w", resp.StatusCode, err)
	}
	return nil
}

func (c *Client) GetJSON(urlPath string, target any) error {
	req, err := http.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return err
	}
	return c.DoJSON(req, target)
}

func (c *Client) PutJSON(urlPath string, input any, target any) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, urlPath, bytes.NewReader(data))
	if err != nil {
		return err
	}
	return c.DoJSON(req, target)
}

func (c *Client) PostJSON(urlPath string, input any, target any) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, urlPath, bytes.NewReader(data))
	if err != nil {
		return err
	}
	return c.DoJSON(req, target)
}

func (c *Client) DeleteJSON(urlPath string, target any) error {
	req, err := http.NewRequest(http.MethodDelete, urlPath, nil)
	if err != nil {
		return err
	}
	return c.DoJSON(req, target)
}

// HTML Pages / Forms
//

func (c *Client) DoPage(req *http.Request) (string, error) {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("got %d code and failed to read response body: %w", resp.StatusCode, err)
		}

		if !utf8.Valid(body) {
			return "", fmt.Errorf("got %d code and %d bytes of binary data", resp.StatusCode, len(body))
		}
		return "", fmt.Errorf("got %d code and response: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("got %d code and failed to read response body: %w", resp.StatusCode, err)
	}
	return string(body), nil
}

func (c *Client) GetPage(urlPath string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		return "", err
	}
	return c.DoPage(req)
}

func (c *Client) PostPage(urlPath string, input url.Values) (string, error) {
	req, err := http.NewRequest(http.MethodPost, urlPath, strings.NewReader(input.Encode()))
	if err != nil {
		return "", err
	}
	return c.DoPage(req)
}

func (c *Client) PutPage(urlPath string, input url.Values) (string, error) {
	req, err := http.NewRequest(http.MethodPut, urlPath, strings.NewReader(input.Encode()))
	if err != nil {
		return "", err
	}
	return c.DoPage(req)
}

func (c *Client) DeletePage(urlPath string) (string, error) {
	req, err := http.NewRequest(http.MethodDelete, urlPath, nil)
	if err != nil {
		return "", err
	}
	return c.DoPage(req)
}
