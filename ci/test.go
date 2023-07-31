package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"golang.org/x/sync/errgroup"
)

func runTests(ctx context.Context, dc *dagger.Client, srcDir *dagger.Directory, goCache *dagger.CacheVolume, nodeCache *dagger.CacheVolume) error {
	mailpitService := dc.Container().From(mailpitImage).WithExposedPort(1025).WithExposedPort(8025)
	mailpitSMTPAddr, err := mailpitService.Endpoint(ctx, dagger.ContainerEndpointOpts{
		Port: 1025,
	})
	if err != nil {
		return fmt.Errorf("failed to get mailhog SMTP addr: %w", err)
	}
	mailpitAPIAddr, err := mailpitService.Endpoint(ctx, dagger.ContainerEndpointOpts{
		Port: 8025,
	})
	if err != nil {
		return fmt.Errorf("failed to get mailhogAPIAddr: %w", err)
	}

	goContainer := getGoContainer(ctx, dc, goContainerOptions{
		cache:  goCache,
		srcDir: srcDir,
	}).
		WithServiceBinding("mailpit", mailpitService).
		WithEnvVariable("MAILPIT_SMTP_ADDR", mailpitSMTPAddr).
		WithEnvVariable("MAILPIT_API_ADDR", mailpitAPIAddr)

	nodeContainer := dc.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From(nodeImage).
		WithMountedCache("/frontend/node_modules", nodeCache).
		WithWorkdir("/src/frontend").
		WithMountedDirectory("/src", srcDir)

	grp, ctx := errgroup.WithContext(ctx)

	// Run backend tests
	goContainer = goContainer.WithExec([]string{"go", "test", "-v", "./..."})
	grp.Go(func() error {
		_, err := goContainer.ExitCode(ctx)
		return err
	})

	// Run frontend tests
	nodeContainer = nodeContainer.WithExec([]string{"yarn"}).
		WithExec([]string{"yarn", "run", "jest"})
	grp.Go(func() error {
		_, err := nodeContainer.ExitCode(ctx)
		return err
	})

	return grp.Wait()
}
