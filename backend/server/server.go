package server

import (
	"dbpedia-server/handlers"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	router  *gin.Engine
	handler *handlers.Handler
	port    string
}

// Config holds server configuration
type Config struct {
	Port        string
	ReleaseMode bool
}

// NewServer creates a new server instance
func NewServer(handler *handlers.Handler, config Config) *Server {
	// Set Gin mode
	if config.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()

	// Setup CORS middleware
	router.Use(corsMiddleware())

	return &Server{
		router:  router,
		handler: handler,
		port:    config.Port,
	}
}

// SetupRoutes configures all the routes
func (s *Server) SetupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.handler.HealthCheck)

	// API v1 group
	v1 := s.router.Group("/api/v1")
	{
		// SPARQL query endpoint
		v1.POST("/sparql", s.handler.ExecuteSPARQLQuery)

		// Validate SPARQL query endpoint
		v1.POST("/validate", s.handler.ValidateSPARQLQuery)

		// Example queries endpoint
		v1.GET("/examples", s.handler.GetExampleQueries)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	fmt.Printf("Server starting on http://localhost%s\n", s.port)
	fmt.Printf("API endpoints:\n")
	fmt.Printf("  - GET  /health\n")
	fmt.Printf("  - POST /api/v1/sparql\n")
	fmt.Printf("  - POST /api/v1/validate\n")
	fmt.Printf("  - GET  /api/v1/examples\n")

	return s.router.Run(s.port)
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
