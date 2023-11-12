package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
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
		From(goreleaserImage).
		WithMountedCache("/go/pkg", opts.goCache).
		WithWorkdir("/src").
		WithMountedDirectory("/src", opts.srcDir).
		WithDirectory("/src/frontend", nodeContainer.Directory("/src/frontend"))

	dockerImageTag := "zerok/webmentiond:latest"
	if opts.releaseVersion != "" {
		goreleaserContainer = goreleaserContainer.WithEnvVariable("RELEASE_VERSION", opts.releaseVersion)
		dockerImageTag = fmt.Sprintf("zerok/webmentiond:%s", opts.releaseVersion)
	}
	goreleaserContainer = goreleaserContainer.WithExec([]string{"release", "--skip-before", "--skip-publish", "--skip-validate"})

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
		var releasePath string
		if opts.releaseVersion != "" {
			releasePath = opts.releaseVersion
		} else {
			releasePath = fmt.Sprintf("snapshots/%s", opts.commitID)
		}
		_, err := dc.Container().
			From(rcloneImage).
			WithEntrypoint(nil).
			WithEnvVariable("RELEASE_PATH", releasePath).
			WithSecretVariable("AWS_S3_BUCKET", opts.awsS3Bucket).
			WithSecretVariable("AWS_S3_ENDPOINT", opts.awsS3Endpoint).
			WithEnvVariable("GIT_COMMIT_ID", opts.commitID).
			WithSecretVariable("AWS_ACCESS_KEY_ID", opts.awsAccessKeyID).
			WithSecretVariable("AWS_SECRET_ACCESS_KEY", opts.awsSecretAccessKey).
			WithSecretVariable("AWS_S3_REGION", opts.awsS3Region).
			WithDirectory("/src", goreleaserContainer.Directory("/src/dist")).
			WithWorkdir("/src").
			WithExec([]string{"sh", "-c", `rclone config create s3 s3 access_key_id=${AWS_ACCESS_KEY_ID} secret_access_key=${AWS_SECRET_ACCESS_KEY} endpoint=${AWS_S3_ENDPOINT} acl=public-read region=${AWS_S3_REGION} > /dev/null`}).
			WithExec([]string{"sh", "-c", `rclone sync . s3:${AWS_S3_BUCKET}/releases/webmentiond/${RELEASE_PATH}`}).
			Sync(ctx)
		return err
	} else {
		_, err := goreleaserContainer.Directory("/src/dist").Export(ctx, "./dist")
		return err
	}
}
