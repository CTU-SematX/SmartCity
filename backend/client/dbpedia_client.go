package client

import (
	"dbpedia-server/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultEndpoint = "https://dbpedia.org/sparql"
	DefaultTimeout  = 30 * time.Second
)

// DBpediaClient implements the DBpedia client
type DBpediaClient struct {
	endpoint   string
	httpClient *http.Client
}

// NewDBpediaClient creates a new DBpedia client
func NewDBpediaClient(endpoint string) *DBpediaClient {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	return &DBpediaClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// Query executes a SPARQL query against DBpedia
func (c *DBpediaClient) Query(sparql string) (*types.SPARQLResponse, error) {
	// Build request URL with query parameters
	params := url.Values{}
	params.Add("query", sparql)
	params.Add("format", "json")

	requestURL := fmt.Sprintf("%s?%s", c.endpoint, params.Encode())

	// Create GET request
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/sparql-results+json")
	req.Header.Set("User-Agent", "DBpedia-Go-Client/1.0")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("dbpedia returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var sparqlResp types.SPARQLResponse
	if err := json.Unmarshal(body, &sparqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &sparqlResp, nil
}
