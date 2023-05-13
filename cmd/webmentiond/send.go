package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zerok/webmentiond/pkg/webmention"
)

func newSendCmd() Command {
	var sendCmd = &cobra.Command{
		Use:   "send SOURCE [TARGET]",
		Short: "Send a mention from source to target",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := logger.WithContext(context.Background())
			failed := false
			if len(args) < 1 {
				return fmt.Errorf("source is required")
			}
			doc, err := webmention.DocumentFromURL(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to load document from URL: %w", err)
			}
			targets := doc.ExternalLinks()
			if len(args) >= 2 {
				targets = []string{args[1]}
			}
			for _, target := range targets {
				mention := webmention.Mention{
					Source: args[0],
					Target: target,
				}
				ep, err := cmd.Flags().GetString("endpoint")
				if err != nil {
					return fmt.Errorf("failed to parse endpoint from flag: %w", err)
				}
				if ep == "" {
					disc := webmention.NewEndpointDiscoverer()
					ep, err = disc.DiscoverEndpoint(ctx, mention.Target)
					if err != nil {
						logger.Warn().Err(err).Msgf("error while looking up endpoint for %s", target)
						continue
					}
					if ep == "" {
						logger.Warn().Err(err).Msgf("%s doesn't expose webmention endpoint", target)
						failed = true
						continue
					}
				}
				sender := webmention.NewSender()
				logger.Info().Msgf("Endpoint: %s", ep)
				if err := sender.Send(ctx, ep, mention); err != nil {
					logger.Error().Err(err).Msgf("Failed to send webmention to %s", target)
				}
			}

			if exitOnFailure, _ := cmd.Flags().GetBool("fail"); failed && exitOnFailure {
				return fmt.Errorf("sending webmentions failed")

			}
			return nil
		},
	}

	sendCmd.Flags().String("endpoint", "", "Endpoint to send the mention to")
	sendCmd.Flags().Bool("fail", false, "Exit with error code if sending a webmention fails")
	return newBaseCommand(sendCmd)
}
