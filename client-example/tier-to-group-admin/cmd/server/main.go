package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"tier-to-group-admin/docs"
	"tier-to-group-admin/internal/api"
	"tier-to-group-admin/internal/service"
	"tier-to-group-admin/internal/storage"
)

// @title           Tier-to-Group Admin API
// @version         1.0
// @description     REST API service for managing tier-to-group mappings in the Open Data Hub Model as a Service (MaaS) project.
// @description     This API provides CRUD operations for managing tiers that map Kubernetes groups to user-defined subscription tiers.

// @contact.name   API Support
// @contact.url    https://github.com/opendatahub-io/maas-billing
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes   http https

func init() {
	// Initialize Swagger docs
	docs.SwaggerInfo.Title = "Tier-to-Group Admin API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "REST API service for managing tier-to-group mappings in the Open Data Hub Model as a Service (MaaS) project."
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}

// @title           Tier-to-Group Admin API
// @version         1.0
// @description     REST API service for managing tier-to-group mappings in the Open Data Hub Model as a Service (MaaS) project.
// @description     This API provides CRUD operations for managing tiers that map Kubernetes groups to user-defined subscription tiers.

// @contact.name   API Support
// @contact.url    https://github.com/opendatahub-io/maas-billing
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @schemes   http https

func main() {
	// Command line flags
	filePath := flag.String("file", "tier-config.yaml", "Path to the tier configuration file")
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Initialize storage
	tierStorage := storage.NewFileTierStorage(*filePath)

	// Initialize service
	tierService := service.NewTierService(tierStorage)

	// Setup router
	router := api.SetupRouter(tierService)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Tier configuration file: %s", *filePath)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
