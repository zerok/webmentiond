package main

import (
	"context"

	"dagger.io/dagger"
)

func runBuildPackages(ctx context.Context, dc *dagger.Client, srcDir *dagger.Directory, goCache *dagger.CacheVolume, nodeCache *dagger.CacheVolume) error {
	// First we build the frontend code which we can then mount into goreleaser:
	nodeContainer := getNodeContainer(ctx, dc, nodeContainerOptions{
		cache:  nodeCache,
		srcDir: srcDir,
	}).
		WithExec([]string{"yarn"}).
		WithExec([]string{"yarn", "run", "webpack", "--mode", "production"})

	goreleaserContainer := dc.Container().
		From("goreleaser/goreleaser").
		WithMountedCache("/go/pkg", goCache).
		WithWorkdir("/src").
		WithMountedDirectory("/src", srcDir).
		WithDirectory("/src/frontend", nodeContainer.Directory("/src/frontend"))
	_, err := goreleaserContainer.WithExec([]string{"build", "--snapshot", "--skip-before"}).ExitCode(ctx)

	// TODO: Also extract these packages and upload them where necessary or do that with goreleaser directly.
	// TODO: Build docker image and upload to the registry
	return err
}
