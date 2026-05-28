package main

import (
	"context"
	"os"

	"shed/internal/app"
	shedfs "shed/internal/fs"
)

func main() {
	code := app.Run(context.Background(), app.Options{
		Args:     os.Args[1:],
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Resolver: shedfs.SelectedFolderResolver{},
		Scanner:  app.EmptyScanner{},
	})
	os.Exit(code)
}
