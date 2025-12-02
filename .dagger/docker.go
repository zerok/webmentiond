package main

import (
	"context"
	"dagger/webmentiond/internal/dagger"
)

func (m *Webmentiond) CreateDockerImage(ctx context.Context, binaries *dagger.Directory, tag string) (string, error) {
	amd64Binary := binaries.File("webmentiond-linux-amd64-musl")
	arm64Binary := binaries.File("webmentiond-linux-arm64-musl")

	amd64Container := dag.
		Container(dagger.ContainerOpts{
			Platform: "linux/amd64",
		}).
		From("alpine:3.22").
		WithExec([]string{"adduser", "-u", "1500", "-h", "/data", "-H", "-D", "webmentiond"}).
		WithUser("webmentiond").
		WithWorkdir("/var/lib/webmentiond").
		WithFile("/bin/webmentiond", amd64Binary).
		WithEntrypoint([]string{"/bin/webmentiond", "serve", "/var/lib/webmentiond/migrations", "--database", "/data/webmentiond.sqlite"})

	arm64Container := dag.
		Container(dagger.ContainerOpts{
			Platform: "linux/arm64",
		}).
		From("alpine:3.22").
		WithExec([]string{"adduser", "-u", "1500", "-h", "/data", "-H", "-D", "webmentiond"}).
		WithUser("webmentiond").
		WithWorkdir("/var/lib/webmentiond").
		WithFile("/bin/webmentiond", arm64Binary).
		WithEntrypoint([]string{"/bin/webmentiond", "serve", "/var/lib/webmentiond/migrations", "--database", "/data/webmentiond.sqlite"})

	variants := []*dagger.Container{
		amd64Container,
		arm64Container,
	}

	// If we are also publishing latest, we need to do that command twice:
	return dag.Container().Publish(ctx, "zerok/webmentiond:"+tag, dagger.ContainerPublishOpts{
		PlatformVariants: variants,
	})
}
