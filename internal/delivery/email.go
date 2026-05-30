package delivery

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/kextant/kextant-agent/internal/config"
	"github.com/kextant/kextant-agent/internal/report"
	"github.com/kextant/kextant-agent/pkg/types"
)

type EmailDelivery struct {
	config *config.Config
	logger *slog.Logger
}

func NewEmailDelivery(cfg *config.Config, logger *slog.Logger) *EmailDelivery {
	return &EmailDelivery{
		config: cfg,
		logger: logger,
	}
}

func (e *EmailDelivery) Send(r *types.Report) error {
	e.logger.Info("sending report via email",
		"cluster", r.ClusterName,
		"recipients", e.config.EmailRecipients)

	htmlBody, err := report.GenerateHTML(r)
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %w", err)
	}

	subject, err := e.renderSubject(r)
	if err != nil {
		return fmt.Errorf("failed to render subject: %w", err)
	}

	return e.deliverWithRetry(subject, htmlBody)
}

func (e *EmailDelivery) renderSubject(r *types.Report) (string, error) {
	tmpl, err := template.New("subject").Parse(e.config.EmailSubject)
	if err != nil {
		return "", err
	}

	data := map[string]string{
		"ClusterName": r.ClusterName,
		"Date":        r.GeneratedAt.Format("Jan 2, 2006"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (e *EmailDelivery) deliverWithRetry(subject, htmlBody string) error {
	delays := []time.Duration{1 * time.Second, 5 * time.Second, 30 * time.Second}

	var lastErr error
	for i, delay := range delays {
		if err := e.sendEmail(subject, htmlBody); err != nil {
			lastErr = err
			e.logger.Error("email delivery failed",
				"attempt", i+1,
				"max_attempts", len(delays),
				"error", err)
			if i < len(delays)-1 {
				time.Sleep(delay)
			}
			continue
		}
		e.logger.Info("successfully delivered report via email")
		return nil
	}

	return fmt.Errorf("email delivery failed after %d attempts: %w", len(delays), lastErr)
}

func (e *EmailDelivery) sendEmail(subject, htmlBody string) error {
	// Build message
	msg := e.buildMessage(subject, htmlBody)

	// Determine auth
	var auth smtp.Auth
	if e.config.SMTPUsername != "" && e.config.SMTPPassword != "" {
		switch e.config.SMTPAuthType {
		case "plain":
			auth = smtp.PlainAuth("", e.config.SMTPUsername, e.config.SMTPPassword, e.config.SMTPHost)
		case "login":
			auth = LoginAuth(e.config.SMTPUsername, e.config.SMTPPassword)
		case "crammd5":
			auth = smtp.CRAMMD5Auth(e.config.SMTPUsername, e.config.SMTPPassword)
		default:
			auth = smtp.PlainAuth("", e.config.SMTPUsername, e.config.SMTPPassword, e.config.SMTPHost)
		}
	}

	addr := fmt.Sprintf("%s:%s", e.config.SMTPHost, e.config.SMTPPort)

	// Handle TLS modes
	switch e.config.SMTPTLSMode {
	case "tls":
		return e.sendWithTLS(addr, auth, msg)
	case "starttls":
		return e.sendWithSTARTTLS(addr, auth, msg)
	default:
		// No TLS
		return smtp.SendMail(addr, auth, e.config.EmailFrom, e.config.EmailRecipients, msg)
	}
}

func (e *EmailDelivery) sendWithTLS(addr string, auth smtp.Auth, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName:         e.config.SMTPHost,
		InsecureSkipVerify: e.config.SMTPTLSSkipVerify,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, e.config.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(e.config.EmailFrom); err != nil {
		return err
	}

	for _, recipient := range e.config.EmailRecipients {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err := w.Write(msg); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func (e *EmailDelivery) sendWithSTARTTLS(addr string, auth smtp.Auth, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName:         e.config.SMTPHost,
		InsecureSkipVerify: e.config.SMTPTLSSkipVerify,
	}

	if err := client.StartTLS(tlsConfig); err != nil {
		return err
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(e.config.EmailFrom); err != nil {
		return err
	}

	for _, recipient := range e.config.EmailRecipients {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err := w.Write(msg); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func (e *EmailDelivery) buildMessage(subject, htmlBody string) []byte {
	headers := make(map[string]string)
	headers["From"] = e.config.EmailFrom
	headers["To"] = strings.Join(e.config.EmailRecipients, ", ")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	return msg.Bytes()
}

// LoginAuth implements LOGIN authentication
type loginAuth struct {
	username string
	password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:", "User Name\x00":
			return []byte(a.username), nil
		case "Password:", "Password\x00":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unknown fromServer: %s", string(fromServer))
		}
	}
	return nil, nil
}
