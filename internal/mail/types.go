package mail

import "time"

type Credentials struct {
	IMAPHost string
	IMAPPort string
	SMTPHost string
	SMTPPort string
	Username string
	Password string
}

type Mailbox struct {
	Name       string
	Delimiter  string
	Flags      []string
	Subscribed bool
}

type MessageHeader struct {
	UID           uint32
	Subject       string
	From          string
	Date          time.Time
	Seen          bool
	HasAttachment bool
}

type Message struct {
	MessageHeader
	To          []string
	CC          []string
	TextBody    string
	HTMLBody    string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Size        int
	Data        []byte
}

type ComposeRequest struct {
	From    string
	To      string
	CC      string
	Subject string
	Body    string
}
