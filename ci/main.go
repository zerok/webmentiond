package main

import (
	"context"
	"os"

	"dagger.io/dagger"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

// TODO: Read this image from an external source
var goImage = "golang:1.24.2-alpine"
var nodeImage = "node:18-alpine"
var alpineImage = "alpine:3.21"
var mailpitImage = "axllent/mailpit:v1.8"
var rcloneImage = "rclone/rclone:1.64"

func main() {
	ctx := context.Background()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx = logger.WithContext(ctx)

	var doBuild bool
	var doTest bool
	var doWebsite bool
	var doPublish bool
	var platforms []string

	pflag.BoolVar(&doBuild, "build", false, "Generate binary package")
	pflag.BoolVar(&doTest, "test", false, "Execute tests")
	pflag.BoolVar(&doWebsite, "website", false, "Build the website")
	pflag.BoolVar(&doPublish, "publish", false, "Publish the generated packages and website")
	pflag.StringSliceVar(&platforms, "platform", []string{"linux/amd64", "linux/arm64"}, "Platforms to generate output for")
	pflag.Parse()

	dc, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		logger.Fatal().Err(err).Msgf("Failed to connect to Dagger Engine")
	}
	defer dc.Close()

	// Register all the environment variables that we'll need throughout the run:
	commitID := requireEnv(ctx, "GIT_COMMIT_ID", doBuild)
	sshPrivateKeySecret := dc.SetSecret("SSH_PRIVATE_KEY", requireEnv(ctx, "SSH_PRIVATE_KEY", doWebsite && doPublish))

	goCache := dc.CacheVolume("go-cache")
	nodeCache := dc.CacheVolume("node-cache")

	srcDir := dc.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"frontend/node_modules", "bin", "data", ".github", "dist"},
	})

	if doTest {
		if err := runTests(ctx, dc, srcDir, goCache, nodeCache); err != nil {
			logger.Fatal().Err(err).Msg("Tests failed")
		}
	}

	if doBuild {
		if err := runBuildPackages(ctx, dc, buildPackageOptions{
			commitID:       commitID,
			srcDir:         srcDir,
			goCache:        goCache,
			nodeCache:      nodeCache,
			publish:        doPublish,
			releaseVersion: os.Getenv("RELEASE_VERSION"),
			platforms:      platforms,
			imageName:      os.Getenv("IMAGE_NAME"),
		}); err != nil {
			logger.Fatal().Err(err).Msg("Package building failed")
		}
	}

	if doWebsite {
		if err := runBuildWebsite(ctx, dc, buildWebsiteOptions{
			srcDir:        srcDir,
			publish:       doPublish,
			sshPrivateKey: sshPrivateKeySecret,
		}); err != nil {
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

func requireEnv(ctx context.Context, name string, conditional bool) string {
	logger := zerolog.Ctx(ctx)
	val := os.Getenv(name)
	if conditional && val == "" {
		logger.Fatal().Msgf("%s not set", name)
	}
	return val
}
