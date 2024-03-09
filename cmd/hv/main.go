package main

import (
	"context"
	"log"
	"os"

	"github.com/robotomize/go-hv/pkg/gohist"
	"github.com/robotomize/go-hv/pkg/gosnap"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	snapper := gosnap.New(homeDir, ".hv", gohist.NewTextMarshaller(), gosnap.WithZSH(), gosnap.WithBASH())
	if err := snapper.Snapshot(context.Background()); err != nil {
		log.Fatal(err)
	}
}
