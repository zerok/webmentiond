package main

import (
	"context"
	"dagger/webmentiond/internal/dagger"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"

	"github.com/google/go-github/v79/github"
)

var goImage = "golang:1.25.5"
var nodeImage = "node:18"
var alpineImage = "alpine:3.22"
var mailpitImage = "axllent/mailpit:v1.8"

type Webmentiond struct{}

func (m *Webmentiond) BuildFrontend(frontendDir *dagger.Directory) *dagger.Directory {
	cache := dag.CacheVolume("frontend-node")
	cacheNodeModules := dag.CacheVolume("frontend-node-modules")
	return dag.Container().From(nodeImage).
		WithMountedCache("/root/.npm", cache).
		WithDirectory("/src", frontendDir, dagger.ContainerWithDirectoryOpts{
			Exclude: []string{"node_modules/**"},
		}).
		WithMountedCache("/src/node_modules", cacheNodeModules).
		WithWorkdir("/src").
		WithExec([]string{"yarn"}).
		WithExec([]string{"yarn", "run", "webpack", "--mode", "production"}).
		Directory("/src")
}

func (m *Webmentiond) TestBackend(ctx context.Context, rootDir *dagger.Directory) (*dagger.Container, error) {

	mailpitService := dag.Container().From(mailpitImage).WithExposedPort(1025).WithExposedPort(8025).AsService()
	mailpitSMTPAddr, err := mailpitService.Endpoint(ctx, dagger.ServiceEndpointOpts{
		Port: 1025,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get mailhog SMTP addr: %w", err)
	}
	mailpitAPIAddr, err := mailpitService.Endpoint(ctx, dagger.ServiceEndpointOpts{
		Port: 8025,
	})
	goCache := dag.CacheVolume("go-pkg")
	return dag.Container().From("golang:1.25.4").
		WithServiceBinding("mailpit", mailpitService).
		WithEnvVariable("MAILPIT_SMTP_ADDR", mailpitSMTPAddr).
		WithEnvVariable("MAILPIT_API_ADDR", mailpitAPIAddr).
		WithMountedCache("/go/pkg", goCache).
		WithDirectory("/src", rootDir).
		WithWorkdir("/src").
		WithExec([]string{"mkdir", "-p", "frontend/css"}).
		WithExec([]string{"mkdir", "-p", "frontend/dist"}).
		WithExec([]string{"touch", "frontend/index.html"}).
		WithExec([]string{"touch", "frontend/demo.html"}).
		WithExec([]string{"touch", "frontend/dist/empty"}).
		WithExec([]string{"touch", "frontend/css/empty"}).
		WithExec([]string{"go", "test", "./..."}).Sync(ctx)
}

func (m *Webmentiond) AttachBinaries(ctx context.Context, token *dagger.Secret, uploadURL string, binaries *dagger.Directory) error {
	githubToken, err := token.Plaintext(ctx)
	if err != nil {
		return err
	}
	repoOwner, repoName, releaseID, err := parseUploadURL(uploadURL)
	if err != nil {
		return err
	}
	client := github.NewClient(nil).WithAuthToken(githubToken)
	entries, err := binaries.Entries(ctx)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		slog.InfoContext(ctx, "uploading file", "file", entry)
		filePath, err := binaries.File(entry).Export(ctx, entry)
		if err != nil {
			return err
		}
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		if _, _, err := client.Repositories.UploadReleaseAsset(ctx, repoOwner, repoName, releaseID, &github.UploadOptions{
			Name: entry,
		}, file); err != nil {
			file.Close()
			return err
		}
		file.Close()
	}
	return nil
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
