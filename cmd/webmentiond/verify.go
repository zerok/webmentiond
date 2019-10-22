package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/zerok/webmentiond/pkg/webmention"
)

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
		client := &http.Client{}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, mention.Source, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		v := webmention.NewVerifier()
		defer resp.Body.Close()
		err = v.Verify(ctx, resp, resp.Body, mention)
		if err == nil {
			logger.Info().Msgf("%s links to %s.", mention.Source, mention.Target)
		} else {
			logger.Fatal().Msgf("No link between %s and %s found.", mention.Source, mention.Target)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
