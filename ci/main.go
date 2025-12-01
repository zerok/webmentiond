package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"dagger.io/dagger"
	"github.com/google/go-github/v79/github"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

// TODO: Read this image from an external source
var goImage = "golang:1.25.4-alpine"
var nodeImage = "node:18-alpine"
var alpineImage = "alpine:3.22"
var mailpitImage = "axllent/mailpit:v1.8"

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

	gh := github.NewClient(nil)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		gh = gh.WithAuthToken(token)
	}

	var owner, repo string
	var releaseID int64
	if uploadURL := os.Getenv("GITHUB_RELEASE_UPLOAD_URL"); uploadURL != "" {
		owner, repo, releaseID, err = parseUploadURL(uploadURL)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to parse upload URL")
		}
	}

	if doTest {
		if err := runTests(ctx, dc, srcDir, goCache, nodeCache); err != nil {
			logger.Fatal().Err(err).Msg("Tests failed")
		}
	}

	if doBuild {
		if err := runBuildPackages(ctx, dc, buildPackageOptions{
			commitID:        commitID,
			srcDir:          srcDir,
			goCache:         goCache,
			nodeCache:       nodeCache,
			publish:         doPublish,
			releaseVersion:  os.Getenv("RELEASE_VERSION"),
			platforms:       platforms,
			imageName:       os.Getenv("IMAGE_NAME"),
			githubClient:    gh,
			githubOwner:     owner,
			githubRepo:      repo,
			githubReleaseID: releaseID,
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

var uploadURLPattern = regexp.MustCompile("https://uploads\\.github\\.com/repos/([^/]+)/([^/]+)/releases/(\\d+)/assets")

func parseUploadURL(v string) (owner string, repo string, releaseID int64, err error) {
	match := uploadURLPattern.FindStringSubmatch(v)
	if len(match) != 4 {
		err = fmt.Errorf("could not parse upload URL")
		return
	}
	owner = match[1]
	repo = match[2]
	rawReleaseID := match[3]
	releaseID, err = strconv.ParseInt(rawReleaseID, 10, 64)
	return
}
