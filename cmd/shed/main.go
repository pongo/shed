package main

import (
	"context"
	"fmt"
	"os"

	"shed/internal/app"
	shedfs "shed/internal/fs"
)

func main() {
	archiveRoot, err := shedfs.UserArchiveRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Archive path unavailable: %v\n", err)
		os.Exit(app.ExitError)
	}

	code := app.Run(context.Background(), app.Options{
		Args:     os.Args[1:],
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Resolver: shedfs.SelectedFolderResolver{},
		Scanner:  shedfs.NewScanner(archiveRoot),
	})
	os.Exit(code)
}
