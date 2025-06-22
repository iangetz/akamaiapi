package akamaiapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	edgegrid "github.com/akamai/AkamaiOPEN-edgegrid-golang/v11/pkg/edgegrid"
)

const (
	maxRetries = 10
	retryDelay = 5 * time.Second
)

type RequestConfig struct {
	EdgercPath    string            // path to .edgerc file
	Section       string            // section in .edgerc
	Path          string            // API endpoint path
	Method        string            // HTTP method (GET, POST, etc.)
	Params        string            // URL query string (without '?')
	Body          interface{}       // request body to be JSON encoded
	Headers       map[string]string // optional headers
	ResponseModel interface{}       // pointer to struct for JSON unmarshal
}

type APIResponse struct {
	StatusCodes []int       // HTTP status codes of each attempt
	Errors      []string    // error messages or response bodies for non-200s
	Body        interface{} // populated ResponseModel on success
}

// DoRequest executes the configured Akamai EdgeGrid API call with retries.
func DoRequest(cfg *RequestConfig) (*APIResponse, error) {
	// Initialize EdgeGrid signer
	config, err := edgegrid.New(
		edgegrid.WithFile(cfg.EdgercPath),
		edgegrid.WithSection(cfg.Section),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize EdgeGrid config: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	var statusCodes []int
	var errorMessages []string
	var respBody []byte

	// Attempt loop
	for i := 0; i < maxRetries; i++ {
		// Prepare body if present
		var bodyReader io.Reader
		if cfg.Body != nil {
			jsonBody, merr := json.Marshal(cfg.Body)
			if merr != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", merr)
			}
			bodyReader = bytes.NewBuffer(jsonBody)
		}

		// Build full URL
		url := fmt.Sprintf("https://%s%s", config.Host, cfg.Path)
		if cfg.Params != "" {
			url += "?" + cfg.Params
		}

		req, rerr := http.NewRequest(cfg.Method, url, bodyReader)
		if rerr != nil {
			return nil, fmt.Errorf("failed to create request: %w", rerr)
		}

		// Set headers
		for k, v := range cfg.Headers {
			req.Header.Set(k, v)
		}
		if cfg.Body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		// Sign request
		config.SignRequest(req)

		// Execute
		resp, derr := client.Do(req)
		if derr != nil {
			statusCodes = append(statusCodes, 0)
			errorMessages = append(errorMessages, derr.Error())
			time.Sleep(retryDelay)
			continue
		}
		defer resp.Body.Close()

		// Read response
		respBody, derr = io.ReadAll(resp.Body)
		statusCodes = append(statusCodes, resp.StatusCode)
		if derr != nil {
			errorMessages = append(errorMessages, derr.Error())
			time.Sleep(retryDelay)
			continue
		}

		// Check status
		if resp.StatusCode == http.StatusOK {
			break
		}
		errorMessages = append(errorMessages, string(respBody))
		time.Sleep(retryDelay)
	}

	// If last attempt failed
	if len(statusCodes) == 0 || statusCodes[len(statusCodes)-1] != http.StatusOK {
		return &APIResponse{StatusCodes: statusCodes, Errors: errorMessages, Body: nil}, nil
	}

	// Unmarshal into provided model
	if cfg.ResponseModel != nil {
		if uerr := json.Unmarshal(respBody, cfg.ResponseModel); uerr != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", uerr)
		}
	}

	return &APIResponse{
		StatusCodes: statusCodes,
		Errors:      errorMessages,
		Body:        cfg.ResponseModel,
	}, nil
}
