package interfaces

import "dbpedia-server/types"

// DBpediaClient defines the interface for interacting with DBpedia
type DBpediaClient interface {
	// Query executes a SPARQL query and returns the results
	Query(sparql string) (*types.SPARQLResponse, error)
}
