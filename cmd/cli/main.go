package main

import (
	"flag"
	"log"

	"github.com/funcman/back-pushing/internal/cli"
)

func main() {
	mapping := flag.String("mapping", "", "Mapping config file path")
	env := flag.String("env", "", "Environment file path")
	flag.Parse()

	if *mapping == "" {
		log.Fatal("--mapping is required")
	}

	if err := cli.Import(*mapping, *env); err != nil {
		log.Fatalf("Import failed: %v", err)
	}
}