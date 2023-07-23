package main

import (
	"context"

	"dagger.io/dagger"
)

func runBuildWebsite(ctx context.Context, dc *dagger.Client, srcDir *dagger.Directory) error {
	container := dc.Container().From("zerok/mkdocs:latest").WithMountedDirectory("/data", srcDir)

	_, err := container.WithExec([]string{"build"}).ExitCode(ctx)
	// TODO: Upload the website
	return err
}
