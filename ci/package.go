package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

type buildPackageOptions struct {
	awsS3Bucket        *dagger.Secret
	awsAccessKeyID     *dagger.Secret
	awsSecretAccessKey *dagger.Secret
	publish            bool
	goCache            *dagger.CacheVolume
	nodeCache          *dagger.CacheVolume
	srcDir             *dagger.Directory
}

func runBuildPackages(ctx context.Context, dc *dagger.Client, opts buildPackageOptions) error {
	// First we build the frontend code which we can then mount into goreleaser:
	nodeContainer := getNodeContainer(ctx, dc, nodeContainerOptions{
		cache:  opts.nodeCache,
		srcDir: opts.srcDir,
	}).
		WithExec([]string{"yarn"}).
		WithExec([]string{"yarn", "run", "webpack", "--mode", "production"})

	goreleaserContainer := dc.Container().
		From("goreleaser/goreleaser").
		WithMountedCache("/go/pkg", opts.goCache).
		WithWorkdir("/src").
		WithMountedDirectory("/src", opts.srcDir).
		WithDirectory("/src/frontend", nodeContainer.Directory("/src/frontend")).
		WithExec([]string{"release", "--snapshot", "--skip-before"})

	commitID := os.Getenv("GIT_COMMIT_ID")
	if commitID == "" {
		return fmt.Errorf("no GIT_COMMIT_ID set")
	}
	syncURL := fmt.Sprintf("s3://%s/releases/webmentiond/snapshots/%s", os.Getenv("AWS_S3_BUCKET"), commitID)
	dockerImageTag := "zerok/webmentiond:latest"

	dockerContainer := dc.Container(dagger.ContainerOpts{
		Platform: "linux/amd64",
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
		WithFile("/usr/local/bin/webmentiond", goreleaserContainer.File("/src/dist/default_linux_amd64_v1/webmentiond")).
		WithUser("webmentiond").
		WithWorkdir("/var/lib/webmentiond").
		WithEntrypoint([]string{"/usr/local/bin/webmentiond", "serve", "--database-migrations", "/var/lib/webmentiond/migrations", "--database", "/data/webmentiond.sqlite"})

	if opts.publish {
		if _, err := dockerContainer.Publish(ctx, dockerImageTag); err != nil {
			return err
		}
		_, err := dc.Container().
			From(awsCLIImage).
			WithSecretVariable("AWS_S3_BUCKET", opts.awsS3Bucket).
			WithSecretVariable("AWS_ACCESS_KEY_ID", opts.awsAccessKeyID).
			WithSecretVariable("AWS_SECRET_ACCESS_KEY", opts.awsSecretAccessKey).
			WithDirectory("/src", goreleaserContainer.Directory("/src/dist")).
			WithWorkdir("/src").
			WithExec([]string{"s3", "sync", ".", syncURL, "--endpoint-url", "https://ams3.digitaloceanspaces.com"}).
			ExitCode(ctx)
		return err
	} else {
		_, err := goreleaserContainer.Directory("/src/dist").Export(ctx, "./dist")
		return err
	}
}
