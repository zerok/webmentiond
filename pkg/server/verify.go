package server

import (
	"context"
	"database/sql"
	"time"

	"github.com/rs/zerolog"
	"github.com/zerok/webmentiond/pkg/webmention"
)

// VerifyNextMention tries to take the next pending mention from the
// database and tries to verify it.
func (srv *Server) VerifyNextMention(ctx context.Context) (bool, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("Checking for new mentions.")
	tx, err := srv.cfg.Database.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	m := Mention{}
	if err := tx.QueryRowContext(ctx, "SELECT id, source, target FROM webmentions WHERE status = ? LIMIT 1", MentionStatusNew).Scan(&m.ID, &m.Source, &m.Target); err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	newStatus := MentionStatusVerified
	mention := webmention.Mention{
		Source: m.Source,
		Target: m.Target,
	}
	if err := webmention.Verify(ctx, &mention); err != nil {
		logger.Error().Err(err).Msgf("Failed to verify %s", m.Source)
		newStatus = MentionStatusInvalid
	}
	logger.Debug().Msgf("title: %s", mention.Title)
	if _, err := tx.ExecContext(ctx, "UPDATE webmentions SET status = ? , title = ? WHERE id = ?", newStatus, mention.Title, m.ID); err != nil {
		tx.Rollback()
		return true, err
	} else {
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return true, err
		}
		logger.Info().Msgf("%s -> %s verified", m.Source, m.Target)
		srv.UpdateGlobalMetrics(ctx)
		return true, nil
	}
}

func (srv *Server) StartVerifier(ctx context.Context) {
	logger := zerolog.Ctx(ctx)
	go func() {
		ticker := time.NewTicker(time.Second * 10)
	loop:
		for {
			select {
			case <-ticker.C:
				if _, err := srv.VerifyNextMention(ctx); err != nil {
					logger.Error().Err(err).Msg("Failed to process mention")
				}
				continue loop
			case <-ctx.Done():
				return
			}
		}
	}()
}
