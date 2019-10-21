package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zerok/webmentiond/pkg/webmention"
)

var sendCmd = &cobra.Command{
	Use:   "send SOURCE TARGET",
	Short: "Send a mention from source to target",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := logger.WithContext(context.Background())
		if len(args) < 2 {
			return fmt.Errorf("source and target required")
		}
		mention := webmention.Mention{
			Source: args[0],
			Target: args[1],
		}

		disc := webmention.NewEndpointDiscoverer()
		sender := webmention.NewSender()
		ep, err := disc.DiscoverEndpoint(ctx, mention.Target)
		if err != nil {
			return fmt.Errorf("error while looking for endpoint: %w", err)
		}
		if ep == "" {
			return fmt.Errorf("%s doesn't expose a webmention endpoint", mention.Target)
		}
		logger.Info().Msgf("Discovered endpoint: %s", ep)
		if err := sender.Send(ctx, ep, mention); err != nil {
			return fmt.Errorf("failed to send webmention: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
