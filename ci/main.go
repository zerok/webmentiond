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
var alpineImage = "alpine:3.18"
var mailhogImage = "mailhog/mailhog:latest"
var awsCLIImage = "amazon/aws-cli:2.13.3"

func main() {
	ctx := context.Background()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx = logger.WithContext(ctx)

	var doBuild bool
	var doTest bool
	var doWebsite bool
	var doPublish bool

	pflag.BoolVar(&doBuild, "build", false, "Generate binary package")
	pflag.BoolVar(&doTest, "test", false, "Execute tests")
	pflag.BoolVar(&doWebsite, "website", false, "Build the website")
	pflag.BoolVar(&doPublish, "publish", false, "Publish the generated packages and website")
	pflag.Parse()

	dc, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		logger.Fatal().Err(err).Msgf("Failed to connect to Dagger Engine")
	}
	defer dc.Close()

	// Register all the environment variables that we'll need throughout the run:
	awsS3BucketSecret := requireEnv(ctx, "AWS_S3_BUCKET", doBuild && doPublish)
	awsAccessKeyIDSecret := dc.SetSecret("AWS_ACCESS_KEY_ID", requireEnv(ctx, "AWS_ACCESS_KEY_ID", doBuild && doPublish))
	awsSecretAccessKeySecret := dc.SetSecret("AWS_SECRET_ACCESS_KEY", requireEnv(ctx, "AWS_SECRET_ACCESS_KEY", doBuild && doPublish))
	awsS3Endpoint := requireEnv(ctx, "AWS_S3_ENDPOINT", doBuild && doPublish)
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
			srcDir:             srcDir,
			goCache:            goCache,
			nodeCache:          nodeCache,
			awsS3Bucket:        awsS3BucketSecret,
			awsS3Endpoint:      awsS3Endpoint,
			awsAccessKeyID:     awsAccessKeyIDSecret,
			awsSecretAccessKey: awsSecretAccessKeySecret,
			publish:            doPublish,
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
