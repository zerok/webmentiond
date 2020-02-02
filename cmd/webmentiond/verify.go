package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zerok/webmentiond/pkg/webmention"
)

func newVerifyCmd() Command {

	var verifyCmd = &cobra.Command{
		Use:   "verify SOURCE TARGET",
		Short: "Check if SOURCE contains a link to TARGET",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := logger.WithContext(context.Background())
			if len(args) < 2 {
				return fmt.Errorf("source and target have to be provided")
			}
			mention := webmention.Mention{
				Source: args[0],
				Target: args[1],
			}
			if err := webmention.Verify(ctx, mention); err == nil {
				logger.Info().Msgf("%s links to %s.", mention.Source, mention.Target)
			} else {
				logger.Fatal().Msgf("No link between %s and %s found.", mention.Source, mention.Target)
			}
			return nil
		},
	}
	return newBaseCommand(verifyCmd)
}
