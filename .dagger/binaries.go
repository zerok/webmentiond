package main

import (
	"bytes"
	"context"
	"dagger/webmentiond/internal/dagger"
	"time"
)

// Produce binaries for glibc and musl in the current architecture
func (m *Webmentiond) BuildBinaries(ctx context.Context, rootDir *dagger.Directory, builtFrontend *dagger.Directory, commitID string, version string, arch string) (*dagger.Directory, error) {
	goCache := dag.CacheVolume("go-pkg")

	flags := bytes.Buffer{}
	flags.WriteString("-X 'main.commit=")
	flags.WriteString(commitID)
	flags.WriteString("'")
	flags.WriteString(" -X 'main.version=")
	flags.WriteString(version)
	flags.WriteString("'")
	flags.WriteString(" -X 'main.date=")
	flags.WriteString(time.Now().Format(time.RFC3339))
	flags.WriteString("'")
	finalFlags := flags.String()

	// Prime the cache
	_, err := dag.Container().From("golang:1.25.4").
		WithMountedCache("/go/pkg", goCache).
		WithDirectory("/src", rootDir, dagger.ContainerWithDirectoryOpts{
			Include: []string{"go.mod", "go.sum"},
		}).
		WithWorkdir("/src").
		WithExec([]string{"go", "mod", "download"}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	output := map[string]struct {
		arch          string
		suffix        string
		baseContainer *dagger.Container
		file          *dagger.File
	}{
		"glibc": {
			arch:          arch,
			suffix:        "glibc",
			baseContainer: dag.Container().From("golang:1.25.4"),
		},
		"musl": {
			arch:          arch,
			suffix:        "musl",
			baseContainer: dag.Container().From("golang:1.25.4-alpine").WithExec([]string{"apk", "add", "--no-cache", "gcc", "libc-dev"}),
		},
	}

	for name, config := range output {
		updated := config
		container := config.baseContainer.
			WithMountedCache("/go/pkg", goCache).
			WithDirectory("/src", rootDir).
			WithDirectory("/src/frontend", builtFrontend).
			WithWorkdir("/src/cmd/webmentiond").
			WithEnvVariable("GOARCH", config.arch).
			WithExec([]string{"go", "build", "-o", "../../webmentiond", "-ldflags", finalFlags})
		updated.file = container.File("/src/webmentiond")
		output[name] = updated
	}

	container := dag.Container()
	for _, config := range output {
		container = container.WithFile("/dist/webmentiond-linux-"+config.arch+"-"+config.suffix, config.file)
	}
	return container.Directory("/dist"), nil

}
