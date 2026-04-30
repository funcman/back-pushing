package cli

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/funcman/back-pushing/internal/adapter"
	"github.com/funcman/back-pushing/internal/adapter/csv"
	"github.com/funcman/back-pushing/internal/adapter/json"
	sqladapter "github.com/funcman/back-pushing/internal/adapter/sql"
	"github.com/funcman/back-pushing/internal/mapper"
	"github.com/funcman/back-pushing/internal/storage/memory"
)

func Import(mappingPath string, envPath string) error {
	// 1. Load environment variables
	if envPath != "" {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Warning: .env file not loaded: %v", err)
		}
	}

	// 2. Load mapping config
	cfg, err := mapper.LoadConfig(mappingPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 3. Initialize data source
	var source adapter.DataSource
	switch cfg.Source.Type {
	case "json":
		source = json.NewDataSource(cfg.Source.Path)
	case "csv":
		source = csv.NewDataSource(cfg.Source.Path)
	case "sql":
		dbURL := os.Getenv(cfg.Env["DB_URL"])
		source, err = sqladapter.NewDataSource(dbURL, cfg.Source.Query)
		if err != nil {
			return fmt.Errorf("create SQL adapter: %w", err)
		}
	default:
		return fmt.Errorf("unsupported source type: %s", cfg.Source.Type)
	}
	defer source.Close()

	// 4. Execute mapping
	m := mapper.New(cfg)
	objects, err := m.Map(context.Background(), source)
	if err != nil {
		return fmt.Errorf("map objects: %w", err)
	}

	// 5. Write to storage
	store := memory.NewObjectStore()
	graph := memory.NewGraphStore()
	for _, obj := range objects {
		if err := store.Create(context.Background(), obj.Type, obj.ID, obj.Data); err != nil {
			log.Printf("Warning: create object %s/%s: %v", obj.Type, obj.ID, err)
		}
		for _, link := range obj.Links {
			graph.AddEdge(context.Background(), link.LinkType, link.FromID, link.ToID, link.Props)
		}
	}

	log.Printf("Imported %d objects", len(objects))
	return nil
}