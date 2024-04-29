package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/robotomize/go-hv/internal/mergetool"
	"github.com/robotomize/go-hv/internal/snapshot"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	s := snapshot.New(homeDir, ".hv")
	if err := s.Snapshot(context.Background()); err != nil {
		log.Fatal(err)
	}

	merger := mergetool.New(filepath.Join(homeDir, ".hv"))
	if err := merger.Merge(context.Background()); err != nil {
		log.Fatal(err)
	}

	// parse, err := fileformat.Parse("hist-2024-03-09T20:17:22+03:00.zsh.bak")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// fmt.Println(parse.String())

	// fmt.Println(fileformat.New("bash"))

	// merger := mergefunc.New(filepath.Join(homeDir, ".hv"), hvmarshal.NewTextMarshaller())
	// if err := merger.Merge(context.Background()); err != nil {
	// 	log.Fatal(err)
	// }
}
