package mail

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

func SendMessage(creds *Credentials, req *ComposeRequest) ([]byte, error) {
	addr := fmt.Sprintf("%s:%s", creds.SMTPHost, creds.SMTPPort)

	auth := smtp.PlainAuth("", creds.Username, creds.Password, creds.SMTPHost)

	recipients := parseAddresses(req.To)
	if len(recipients) == 0 {
		return nil, fmt.Errorf("no valid recipients")
	}

	raw := []byte(buildRawMessage(req))

	if err := smtp.SendMail(addr, auth, req.From, recipients, raw); err != nil {
		return nil, fmt.Errorf("smtp send: %w", err)
	}

	return raw, nil
}

func parseAddresses(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func buildRawMessage(req *ComposeRequest) string {
	var sb strings.Builder

	sb.WriteString("From: " + req.From + "\r\n")
	sb.WriteString("To: " + req.To + "\r\n")
	if req.CC != "" {
		sb.WriteString("Cc: " + req.CC + "\r\n")
	}
	sb.WriteString("Subject: " + req.Subject + "\r\n")
	sb.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(req.Body)

	return sb.String()
}
