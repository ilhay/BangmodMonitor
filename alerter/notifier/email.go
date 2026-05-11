package notifier

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Email struct {
	cfg EmailConfig
}

func (e *Email) Send(subject, message string) error {
	host := e.cfg.SMTPHost
	port := e.cfg.SMTPPort
	if port == "" {
		port = "587"
	}
	addr := net.JoinHostPort(host, port)

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", e.cfg.From),
		fmt.Sprintf("To: %s", e.cfg.To),
		fmt.Sprintf("Subject: [BangmodMonitor] %s", subject),
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		message,
	}, "\r\n")

	tlsCfg := &tls.Config{ServerName: host}
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if err := c.StartTLS(tlsCfg); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}
	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := c.Mail(e.cfg.From); err != nil {
		return err
	}
	if err := c.Rcpt(e.cfg.To); err != nil {
		return err
	}
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()
	_, err = fmt.Fprint(wc, msg)
	return err
}
