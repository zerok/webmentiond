package main

import (
	"context"
	"dagger/webmentiond/internal/dagger"
)

type buildWebsiteOptions struct {
	publish       bool
	srcDir        *dagger.Directory
	sshPrivateKey *dagger.Secret
}

func (m *Webmentiond) BuildWebsite(
	ctx context.Context,
	sshPrivateKey *dagger.Secret,
	sourceDir *dagger.Directory,
	publish bool,
) error {
	container := dag.Container().
		From("zerok/mkdocs:latest").
		WithMountedDirectory("/data", sourceDir).
		WithExec([]string{"mkdocs", "build"})
	if publish {
		_, err := dag.Container().From(alpineImage).
			WithExec([]string{"apk", "add", "--no-cache", "rsync", "openssh"}).
			WithMountedSecret("/root/.ssh/id_rsa", sshPrivateKey).
			WithDirectory("/src", container.Directory("/data/site")).
			WithWorkdir("/src").
			WithExec([]string{"rsync", "-e", "ssh -o StrictHostKeyChecking=no", "-avz", ".", "www-webmentiondorg@webmentiond.org:/srv/www/webmentiond.org/www/htdocs/"}).
			Sync(ctx)
		return err
	}
	_, err := container.Sync(ctx)
	return err
}
