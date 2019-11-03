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
		ep, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			return fmt.Errorf("failed to parse endpoint from flag: %w", err)
		}
		if ep == "" {
			disc := webmention.NewEndpointDiscoverer()
			ep, err = disc.DiscoverEndpoint(ctx, mention.Target)
			if err != nil {
				return fmt.Errorf("error while looking for endpoint: %w", err)
			}
			if ep == "" {
				return fmt.Errorf("%s doesn't expose a webmention endpoint", mention.Target)
			}
		}
		sender := webmention.NewSender()
		logger.Info().Msgf("Endpoint: %s", ep)
		if err := sender.Send(ctx, ep, mention); err != nil {
			return fmt.Errorf("failed to send webmention: %w", err)
		}
		return nil
	},
}

func init() {
	sendCmd.Flags().String("endpoint", "", "Endpoint to send the mention to")
	rootCmd.AddCommand(sendCmd)
}
