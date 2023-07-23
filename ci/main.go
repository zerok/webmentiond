package main

import (
	"context"
	"os"

	"dagger.io/dagger"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

var goImage = "golang:1.20.6-alpine"
var nodeImage = "node:18-alpine"
var mailhogImage = "mailhog/mailhog:latest"

func main() {
	ctx := context.Background()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx = logger.WithContext(ctx)

	var doBuild bool
	var doTest bool
	var doWebsite bool

	pflag.BoolVar(&doBuild, "build", false, "Generate binary package")
	pflag.BoolVar(&doTest, "test", false, "Execute tests")
	pflag.BoolVar(&doWebsite, "website", false, "Build the website")
	pflag.Parse()

	dc, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		logger.Fatal().Err(err).Msgf("Failed to connect to Dagger Engine")
	}
	defer dc.Close()

	goCache := dc.CacheVolume("go-cache")
	nodeCache := dc.CacheVolume("node-cache")

	srcDir := dc.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"frontend/node_modules", "bin", "data", ".git", ".github"},
	})

	if doTest {
		if err := runTests(ctx, dc, srcDir, goCache, nodeCache); err != nil {
			logger.Fatal().Err(err).Msg("Tests failed")
		}
	}

	if doBuild {
		if err := runBuildPackages(ctx, dc, srcDir, goCache, nodeCache); err != nil {
			logger.Fatal().Err(err).Msg("Package building failed")
		}
	}

	if doWebsite {
		if err := runBuildWebsite(ctx, dc, srcDir); err != nil {
			logger.Fatal().Err(err).Msg("Website building failed")
		}
	}
}

type goContainerOptions struct {
	cache    *dagger.CacheVolume
	platform dagger.Platform
	srcDir   *dagger.Directory
}

type nodeContainerOptions struct {
	cache  *dagger.CacheVolume
	srcDir *dagger.Directory
}

func getGoContainer(ctx context.Context, dc *dagger.Client, opts goContainerOptions) *dagger.Container {
	container := dc.Container(dagger.ContainerOpts{Platform: opts.platform}).
		From(goImage).
		WithWorkdir("/src").
		WithExec([]string{"apk", "add", "--no-cache", "gcc", "libc-dev", "git", "sqlite-dev"})
	if opts.cache != nil {
		container = container.WithMountedCache("/go/pkg", opts.cache)
	}
	if opts.srcDir != nil {
		container = container.WithMountedDirectory("/src", opts.srcDir)
	}
	return container
}

func getNodeContainer(ctx context.Context, dc *dagger.Client, opts nodeContainerOptions) *dagger.Container {
	container := dc.Container().
		From(nodeImage).
		WithWorkdir("/src/frontend")
	if opts.srcDir != nil {
		container = container.WithMountedDirectory("/src", opts.srcDir)
	}
	if opts.cache != nil {
		container = container.WithMountedCache("/src/frontend/node_modules", opts.cache)
	}
	return container
}
