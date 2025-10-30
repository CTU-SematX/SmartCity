package handlers

import (
	"bytes"
	"dbpedia-server/interfaces"
	"dbpedia-server/types"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// MockDBpediaClient is a mock implementation of the DBpediaClient interface
type MockDBpediaClient struct {
	QueryFunc func(sparql string) (*types.SPARQLResponse, error)
}

func (m *MockDBpediaClient) Query(sparql string) (*types.SPARQLResponse, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(sparql)
	}
	return &types.SPARQLResponse{}, nil
}

func setupTestHandler() *Handler {
	mockClient := &MockDBpediaClient{
		QueryFunc: func(sparql string) (*types.SPARQLResponse, error) {
			return &types.SPARQLResponse{
				Head: types.Head{
					Vars: []string{"s", "p", "o"},
				},
				Results: types.Results{
					Bindings: []map[string]types.Binding{},
				},
			}, nil
		},
	}
	return NewHandler(mockClient)
}

func TestValidateSPARQLQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid SELECT query",
			query:       "SELECT ?s ?p ?o WHERE { ?s ?p ?o } LIMIT 10",
			shouldError: false,
		},
		{
			name:        "Valid CONSTRUCT query",
			query:       "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o } LIMIT 10",
			shouldError: false,
		},
		{
			name:        "Valid ASK query",
			query:       "ASK WHERE { ?s ?p ?o }",
			shouldError: false,
		},
		{
			name:        "Empty query",
			query:       "",
			shouldError: true,
			errorMsg:    "query cannot be empty",
		},
		{
			name:        "Invalid keyword",
			query:       "INVALID QUERY",
			shouldError: true,
			errorMsg:    "query must start with a valid SPARQL keyword",
		},
		{
			name:        "SELECT without WHERE",
			query:       "SELECT ?s ?p ?o LIMIT 10",
			shouldError: true,
			errorMsg:    "SELECT query must contain a WHERE clause",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSPARQLQuery(tt.query)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateSPARQLQueryEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()
	router := gin.New()
	router.POST("/validate", handler.ValidateSPARQLQuery)

	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedValid  bool
	}{
		{
			name:           "Valid query",
			query:          "SELECT ?s WHERE { ?s ?p ?o } LIMIT 10",
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name:           "Invalid query",
			query:          "NOT VALID",
			expectedStatus: http.StatusOK,
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := types.SPARQLRequest{Query: tt.query}
			jsonBody, _ := json.Marshal(reqBody)

			req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			var response types.ValidateResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if response.Valid != tt.expectedValid {
				t.Errorf("Expected valid=%v but got valid=%v", tt.expectedValid, response.Valid)
			}
		})
	}
}

func TestExecuteSPARQLQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()
	router := gin.New()
	router.POST("/sparql", handler.ExecuteSPARQLQuery)

	t.Run("Valid query execution", func(t *testing.T) {
		reqBody := types.SPARQLRequest{Query: "SELECT ?s WHERE { ?s ?p ?o } LIMIT 10"}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/sparql", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d but got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Invalid query rejected", func(t *testing.T) {
		reqBody := types.SPARQLRequest{Query: "INVALID"}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/sparql", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d but got %d", http.StatusBadRequest, w.Code)
		}

		var response types.ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Error != "invalid_sparql" {
			t.Errorf("Expected error 'invalid_sparql' but got '%s'", response.Error)
		}
	})
}

// Ensure MockDBpediaClient implements the interface
var _ interfaces.DBpediaClient = (*MockDBpediaClient)(nil)
