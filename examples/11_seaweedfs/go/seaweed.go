package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Client talks to a SeaweedFS master+volume via HTTP.
type Client struct {
	MasterURL  string
	ProxyBase  string // if set, volume URLs route through this proxy prefix
	HTTPClient *http.Client
}

func NewClient(masterURL string) *Client {
	return &Client{
		MasterURL:  masterURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// VolumeFileURL returns the URL for a file on a volume server.
// In proxy mode (WASM), it routes through the proxy prefix.
func (c *Client) VolumeFileURL(volumeURL, fid string) string {
	if c.ProxyBase != "" {
		return fmt.Sprintf("%s/%s/%s", c.ProxyBase, volumeURL, fid)
	}
	return fmt.Sprintf("http://%s/%s", volumeURL, fid)
}

// AssignResult is the response from POST /dir/assign.
type AssignResult struct {
	Fid       string `json:"fid"`
	URL       string `json:"url"`
	PublicURL string `json:"publicUrl"`
	Count     int    `json:"count"`
}

// Ping checks if the master server is reachable.
func (c *Client) Ping() error {
	resp, err := c.HTTPClient.Get(c.MasterURL + "/dir/status")
	if err != nil {
		return fmt.Errorf("master unreachable: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("master returned status %d", resp.StatusCode)
	}
	return nil
}

// Status returns raw JSON from /dir/status for display.
func (c *Client) Status() (map[string]any, error) {
	resp, err := c.HTTPClient.Get(c.MasterURL + "/dir/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// Assign requests a file ID from the master.
func (c *Client) Assign() (*AssignResult, error) {
	resp, err := c.HTTPClient.Post(c.MasterURL+"/dir/assign", "", nil)
	if err != nil {
		return nil, fmt.Errorf("assign failed: %w", err)
	}
	defer resp.Body.Close()
	var result AssignResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Fid == "" {
		return nil, fmt.Errorf("assign returned empty fid")
	}
	return &result, nil
}

// Upload stores data on the volume server.
func (c *Client) Upload(volumeURL, fid, filename string, data []byte) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	w.Close()

	url := c.VolumeFileURL(volumeURL, fid)
	resp, err := c.HTTPClient.Post(url, w.FormDataContentType(), &buf)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return fmt.Errorf("upload returned status %d", resp.StatusCode)
	}
	return nil
}

// Download retrieves data from the volume server.
func (c *Client) Download(volumeURL, fid string) ([]byte, error) {
	url := c.VolumeFileURL(volumeURL, fid)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// Delete removes a file from the volume server.
func (c *Client) Delete(volumeURL, fid string) error {
	url := c.VolumeFileURL(volumeURL, fid)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
