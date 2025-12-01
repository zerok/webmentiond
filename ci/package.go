package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"dagger.io/dagger"
	"github.com/google/go-github/v79/github"
	"golang.org/x/sync/errgroup"
)

type buildPackageOptions struct {
	commitID           string
	awsS3Bucket        *dagger.Secret
	awsS3Endpoint      *dagger.Secret
	awsS3Region        *dagger.Secret
	awsAccessKeyID     *dagger.Secret
	awsSecretAccessKey *dagger.Secret
	publish            bool
	goCache            *dagger.CacheVolume
	nodeCache          *dagger.CacheVolume
	srcDir             *dagger.Directory
	releaseVersion     string
	platforms          []string
	imageName          string
	githubClient       *github.Client
	githubOwner        string
	githubRepo         string
	githubReleaseID    int64
}

func runBuildPackages(ctx context.Context, dc *dagger.Client, opts buildPackageOptions) error {
	// First we build the frontend code which we can then mount into the final images:
	nodeContainer := getNodeContainer(ctx, dc, nodeContainerOptions{
		cache:  opts.nodeCache,
		srcDir: opts.srcDir,
	}).
		WithExec([]string{"yarn"}).
		WithExec([]string{"yarn", "run", "webpack", "--mode", "production"})

	// Now build the binary for all platforms
	targetPlatforms := opts.platforms
	buildContainers := map[string]*dagger.Container{}

	flags := bytes.Buffer{}
	flags.WriteString("-X 'main.commit=")
	flags.WriteString(opts.commitID)
	flags.WriteString("'")
	flags.WriteString(" -X 'main.version=")
	flags.WriteString(opts.releaseVersion)
	flags.WriteString("'")
	flags.WriteString(" -X 'main.date=")
	flags.WriteString(time.Now().Format(time.RFC3339))
	flags.WriteString("'")

	for _, platform := range targetPlatforms {
		container := getGoContainer(ctx, dc, goContainerOptions{
			cache:    opts.goCache,
			platform: dagger.Platform(platform),
			srcDir:   opts.srcDir,
		}).
			WithDirectory("./frontend", nodeContainer.Directory(".")).
			WithExec([]string{"go", "build", "-o", "webmentiond", "-ldflags", flags.String(), "./cmd/webmentiond"})
		buildContainers[platform] = container
	}

	// These now need to used in the target containers that are eventually
	// published:
	variants := make([]*dagger.Container, 0, len(targetPlatforms))
	for platform, buildContainer := range buildContainers {
		dockerContainer := dc.Container(dagger.ContainerOpts{
			Platform: dagger.Platform(platform),
		}).
			From(alpineImage).
			WithExec([]string{"apk", "add", "--no-cache", "sqlite-dev"}).
			WithExec([]string{"adduser", "-u", "1500", "-h", "/data", "-H", "-D", "webmentiond"}).
			WithExec([]string{"mkdir", "-p", "/var/lib/webmentiond/frontend"}).
			WithDirectory("/var/lib/webmentiond/migrations", opts.srcDir.Directory("pkg/server/migrations")).
			WithDirectory("/var/lib/webmentiond/frontend/dist", nodeContainer.Directory("/src/frontend/dist")).
			WithDirectory("/var/lib/webmentiond/frontend/css", nodeContainer.Directory("/src/frontend/css")).
			WithFile("/var/lib/webmentiond/frontend/index.html", nodeContainer.File("/src/frontend/index.html")).
			WithFile("/var/lib/webmentiond/frontend/demo.html", nodeContainer.File("/src/frontend/demo.html")).
			WithFile("/usr/local/bin/webmentiond", buildContainer.File("/src/webmentiond")).
			WithUser("webmentiond").
			WithWorkdir("/var/lib/webmentiond").
			WithEntrypoint([]string{"/usr/local/bin/webmentiond", "serve", "--database-migrations", "/var/lib/webmentiond/migrations", "--database", "/data/webmentiond.sqlite"})
		variants = append(variants, dockerContainer)
	}

	imageName := opts.imageName
	imageTag := "main-" + opts.commitID
	if opts.releaseVersion != "" {
		imageTag = opts.releaseVersion
	}
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)
	binaryMapping := map[dagger.Platform]string{
		"linux/amd64": "webmentiond-linux-amd64-musl",
		"linux/arm64": "webmentiond-linux-arm64-musl",
	}

	if opts.publish {
		grp, ctx := errgroup.WithContext(ctx)
		grp.Go(func() error {
			_, err := dc.Container().Publish(ctx, fullImageName, dagger.ContainerPublishOpts{
				PlatformVariants: variants,
			})
			return err
		})
		// If we are running in the context of a release, we also need to export the produced binaries
		for _, variant := range variants {
			grp.Go(func() error {
				platform, err := variant.Platform(ctx)
				if err != nil {
					return err
				}
				filename, ok := binaryMapping[platform]
				if !ok {
					return fmt.Errorf("unknown platform %s", platform)
				}
				if _, err := variant.File("/usr/local/bin/webmentiond").Export(ctx, filename); err != nil {
					return err
				}
				if opts.githubReleaseID != 0 {
					fp, err := os.Open(filename)
					if err != nil {
						return err
					}
					defer fp.Close()
					if _, _, err := opts.githubClient.Repositories.UploadReleaseAsset(ctx, opts.githubOwner, opts.githubRepo, opts.githubReleaseID, &github.UploadOptions{
						Name: filename,
					}, fp); err != nil {
						return err
					}
				}
				return nil
			})
		}
		return grp.Wait()
	}

	// If there is no publishing, then we should at least build the images
	for _, variant := range variants {
		_, err := variant.Sync(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
