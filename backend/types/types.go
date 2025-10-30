package types

// SPARQLRequest represents the request body for SPARQL queries
type SPARQLRequest struct {
	Query string `json:"query" binding:"required"`
}

// SPARQLResponse represents the response from DBpedia
type SPARQLResponse struct {
	Head    Head    `json:"head"`
	Results Results `json:"results"`
}

// Head represents the head section of a SPARQL response
type Head struct {
	Link []string `json:"link"`
	Vars []string `json:"vars"`
}

// Results represents the results section of a SPARQL response
type Results struct {
	Bindings []map[string]Binding `json:"bindings"`
	Distinct bool                 `json:"distinct"`
	Ordered  bool                 `json:"ordered"`
}

// Binding represents a single binding in SPARQL results
type Binding struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Lang     string `json:"xml:lang,omitempty"`
	Datatype string `json:"datatype,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ExampleQuery represents an example SPARQL query
type ExampleQuery struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Query       string `json:"query"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// ValidateResponse represents the validation response
type ValidateResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}
