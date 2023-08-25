package main

import (
	"context"

	"dagger.io/dagger"
)

type buildWebsiteOptions struct {
	publish       bool
	srcDir        *dagger.Directory
	sshPrivateKey *dagger.Secret
}

func runBuildWebsite(ctx context.Context, dc *dagger.Client, opts buildWebsiteOptions) error {
	container := dc.Container().From("zerok/mkdocs:latest").WithMountedDirectory("/data", opts.srcDir).WithExec([]string{"build"})
	if opts.publish {
		_, err := dc.Container().From(alpineImage).
			WithExec([]string{"apk", "add", "--no-cache", "rsync", "openssh"}).
			WithMountedSecret("/root/.ssh/id_rsa", opts.sshPrivateKey).
			WithDirectory("/src", container.Directory("/data/site")).
			WithWorkdir("/src").
			WithExec([]string{"rsync", "-e", "ssh -o StrictHostKeyChecking=no", "-avz", ".", "www-webmentiondorg@webmentiond.org:/srv/www/webmentiond.org/www/htdocs/"}).
			Sync(ctx)
		return err
	} else {
		_, err := container.Sync(ctx)
		return err
	}
}
