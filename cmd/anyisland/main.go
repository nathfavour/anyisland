package main

import (
	"fmt"
	"log"

	"github.com/nathfavour/anyisland/internal/pal"
)

func main() {
	sys, err := pal.New()
	if err != nil {
		log.Fatalf("failed to initialize platform abstraction layer: %v", err)
	}

	if err := sys.InitFolders(); err != nil {
		log.Fatalf("failed to initialize island folders: %v", err)
	}

	fmt.Printf("Anyisland initialized at %s\n", sys.GetIslandDir())
}
