package server

import (
	"context"
	"fmt"

	"github.com/zerok/webmentiond/pkg/webmention"
)

func (srv *Server) sendNotificationMail(ctx context.Context, mention webmention.Mention, status string) error {
	return srv.mailer.SendMail(ctx, srv.cfg.MailFrom, srv.cfg.Auth.AdminEmails, "Mention verified", fmt.Sprintf("Source: <%s>\nTarget: <%s>\nNew status: %s\n\nGo to <%s/ui/> for details.", mention.Source, mention.Target, status, srv.cfg.PublicURL))
}
