package mail

import "time"

// Credentials holds the user's mail server connection details.
type Credentials struct {
	IMAPHost string
	IMAPPort string
	SMTPHost string
	SMTPPort string
	Username string
	Password string
}

// Mailbox represents an IMAP folder.
type Mailbox struct {
	Name       string
	Delimiter  string
	Flags      []string
	Subscribed bool
}

// MessageHeader is a lightweight summary shown in the inbox list.
type MessageHeader struct {
	UID           uint32
	Subject       string
	From          string
	Date          time.Time
	Seen          bool
	HasAttachment bool
}

// Message is the full message including body parts.
type Message struct {
	MessageHeader
	To          []string
	CC          []string
	ReplyTo     string
	TextBody    string
	HTMLBody    string
	Attachments []Attachment
}

// Attachment holds metadata and raw bytes for a mail attachment.
type Attachment struct {
	Filename    string
	ContentType string
	Size        int
	Data        []byte
}

// ComposeRequest is the data needed to send an outbound email.
type ComposeRequest struct {
	From    string
	To      string
	CC      string
	Subject string
	Body    string
}
