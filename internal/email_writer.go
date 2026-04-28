package internal

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/quickybrains/copirator/internal/log"
	"github.com/wneessen/go-mail"
)

type EmailWriter struct {
	cfg      EmailOutputConfig
	fileName string
	username string

	cl     *mail.Client
	logger log.Logger
}

func NewEmailWriter(cfg EmailOutputConfig, logger log.Logger) (*EmailWriter, error) {
	fileName := "backup.zip"
	if cfg.FileName != "" {
		fileName = cfg.FileName
	}

	fileName = enrichFileName(fileName)

	username := os.Getenv("EMAIL_USERNAME")
	pwd := os.Getenv("EMAIL_PASSWORD")

	if username == "" || pwd == "" {
		return nil, fmt.Errorf("auth credentials are required")
	}

	client, err := mail.NewClient(
		cfg.Addr,
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(username),
		mail.WithPassword(pwd),
	)
	if err != nil {
		return nil, fmt.Errorf("mail new client: %w", err)
	}

	return &EmailWriter{
		cfg:      cfg,
		fileName: fileName,
		username: username,
		cl:       client,
		logger:   logger.With("component", "email_writer"),
	}, nil
}

func (ew *EmailWriter) Write(data []byte) error {
	if !ew.cfg.Enabled {
		return nil
	}

	r := bytes.NewReader(data)
	msg := mail.NewMsg()

	err := msg.From(ew.username)
	if err != nil {
		return fmt.Errorf("add `from` addr: %w", err)
	}

	err = msg.To(ew.cfg.ToAddr)
	if err != nil {
		return fmt.Errorf("add `to` addr: %w", err)
	}
	msg.Subject(ew.cfg.Subj)

	err = msg.EmbedReader(ew.fileName, r, mail.WithFileContentType(mail.TypeTextPlain))
	if err != nil {
		return fmt.Errorf("embed zip readed: %w", err)
	}

	ew.logger.
		Info().
		String("from", ew.username).
		String("to", ew.cfg.ToAddr).
		String("subject", ew.cfg.Subj).
		Print("Sending email with backup")

	const defaultDialTimeout = 3 * time.Second
	diatTimeout := max(defaultDialTimeout, ew.cfg.DialTimeout)

	ctx, cancelFn := context.WithTimeout(context.Background(), diatTimeout)
	defer cancelFn()

	err = ew.cl.DialAndSendWithContext(ctx, msg)
	if err != nil {
		return fmt.Errorf("dial and send: %w", err)
	}

	return nil
}

func enrichFileName(name string) string {
	var suffix string

	idx := strings.LastIndex(name, ".")
	if idx != -1 {
		suffix = name[idx:]
		name = name[:idx]

	}

	name += " " + time.Now().Format(time.DateTime)

	return name + suffix
}
