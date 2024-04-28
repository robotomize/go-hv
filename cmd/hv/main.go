package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/robotomize/go-hv/internal/enc"
	"github.com/robotomize/go-hv/internal/fileformat"
	"github.com/robotomize/go-hv/internal/mergefunc"
	"github.com/robotomize/go-hv/internal/snapshot"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	snapper := snapshot.New(homeDir, ".hv", enc.NewTextMarshaller(), snapshot.WithZSH(), snapshot.WithBASH())
	if err := snapper.Snapshot(context.Background()); err != nil {
		log.Fatal(err)
	}
	parse, err := fileformat.Parse("hist-2024-03-09T20:17:22+03:00.zsh.bak")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(parse.String())

	// fmt.Println(fileformat.New("bash"))

	merger := mergefunc.New(filepath.Join(homeDir, ".hv"), enc.NewTextMarshaller())
	if err := merger.Merge(context.Background()); err != nil {
		log.Fatal(err)
	}
}
