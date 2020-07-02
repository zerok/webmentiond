package server

import (
	"context"
	"database/sql"
	"time"

	"github.com/rs/zerolog"
	"github.com/zerok/webmentiond/pkg/policies"
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
	// The last verification must be at least a minute in the past
	valid_last_verification := time.Now()
	if srv.cfg.VerificationTimeoutDuration != 0 {
		valid_last_verification = valid_last_verification.Add(-1 * srv.cfg.VerificationTimeoutDuration)
	} else {
		valid_last_verification = valid_last_verification.Add(time.Second)
	}
	if err := tx.QueryRowContext(ctx, "SELECT id, source, target FROM webmentions WHERE status = ? AND (verified_at = '' OR verified_at) < ? LIMIT 1", MentionStatusNew, valid_last_verification.Format(time.RFC3339)).Scan(&m.ID, &m.Source, &m.Target); err != nil {
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
	if err := webmention.Verify(ctx, &mention, func(c *webmention.VerifyOptions) {
		c.MaxRedirects = srv.cfg.VerificationMaxRedirects
	}); err != nil {
		newStatus = MentionStatusInvalid
	}
	if srv.cfg.Policies != nil {
		if newStatus == MentionStatusVerified {
			// Check if we can skip the manual approval process for this URL:
			if srv.cfg.Policies.DetermineForURL(mention.Source) == policies.APPROVE {
				logger.Info().Msgf("%s -> %s auto-approved", mention.Source, mention.Target)
				newStatus = MentionStatusApproved
			}
		}
	}
	if len(mention.Content) > 500 {
		mention.Content = mention.Content[0:497] + "…"
	}
	logger.Debug().Msgf("title: %s", mention.Title)
	if _, err := tx.ExecContext(ctx, "UPDATE webmentions SET status = ? , title = ? , verified_at = ?, type = ?, content = ?, author_name = ?, rsvp = ? WHERE id = ?", newStatus, mention.Title, time.Now().Format(time.RFC3339), mention.Type, mention.Content, mention.AuthorName, mention.RSVP, m.ID); err != nil {
		tx.Rollback()
		return true, err
	} else {
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return true, err
		}
		logger.Debug().Msgf("%s -> %s checked: %v", m.Source, m.Target, newStatus)
		srv.UpdateGlobalMetrics(ctx)
		if srv.cfg.NotifyOnVerification {
			if err := srv.sendNotificationMail(ctx, mention, newStatus); err != nil {
				logger.Error().Err(err).Msg("Failed to send notification email")
			}
		}
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
