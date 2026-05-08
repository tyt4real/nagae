package mail

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

type IMAPClient struct {
	client *imapclient.Client
	creds  *Credentials
}

func NewIMAPClient(creds *Credentials) (*IMAPClient, error) {
	addr := fmt.Sprintf("%s:%s", creds.IMAPHost, creds.IMAPPort)

	var (
		c   *imapclient.Client
		err error
	)

	if creds.IMAPPort == "993" {
		c, err = imapclient.DialTLS(addr, nil)
	} else {
		c, err = imapclient.DialStartTLS(addr, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("dial imap: %w", err)
	}

	if err := c.Login(creds.Username, creds.Password).Wait(); err != nil {
		c.Close()
		return nil, fmt.Errorf("imap login: %w", err)
	}

	return &IMAPClient{client: c, creds: creds}, nil
}

func (ic *IMAPClient) Close() error {
	return ic.client.Logout().Wait()
}

var sentMailboxCandidates = []string{
	"Sent",
	"Sent Items",
	"Sent Messages",
	"INBOX.Sent",
	"[Gmail]/Sent Mail",
}

func (ic *IMAPClient) findSentMailbox() string {
	cmd := ic.client.List("", "*", &imap.ListOptions{
		ReturnStatus: &imap.StatusOptions{NumMessages: true},
	})
	mailboxes, err := cmd.Collect()
	if err != nil {
		return "Sent"
	}

	for _, mb := range mailboxes {
		for _, attr := range mb.Attrs {
			if attr == "\\Sent" {
				return mb.Mailbox
			}
		}
	}

	names := make(map[string]struct{}, len(mailboxes))
	for _, mb := range mailboxes {
		names[mb.Mailbox] = struct{}{}
	}
	for _, candidate := range sentMailboxCandidates {
		if _, ok := names[candidate]; ok {
			return candidate
		}
	}

	return "Sent"
}

func (ic *IMAPClient) AppendToSent(rawMsg []byte) error {
	sentMailbox := ic.findSentMailbox()

	flags := []imap.Flag{imap.FlagSeen}
	size := int64(len(rawMsg))

	appendCmd := ic.client.Append(sentMailbox, size, &imap.AppendOptions{Flags: flags})
	if _, err := appendCmd.Write(rawMsg); err != nil {
		return fmt.Errorf("append write: %w", err)
	}
	if err := appendCmd.Close(); err != nil {
		return fmt.Errorf("append close: %w", err)
	}
	if _, err := appendCmd.Wait(); err != nil {
		_ = err
	}

	return nil
}

func (ic *IMAPClient) ListMailboxes() ([]Mailbox, error) {
	cmd := ic.client.List("", "*", nil)
	mailboxes, err := cmd.Collect()
	if err != nil {
		return nil, fmt.Errorf("list mailboxes: %w", err)
	}

	result := make([]Mailbox, 0, len(mailboxes))
	for _, mb := range mailboxes {
		flags := make([]string, 0, len(mb.Attrs))
		for _, f := range mb.Attrs {
			flags = append(flags, string(f))
		}
		result = append(result, Mailbox{
			Name:      mb.Mailbox,
			Delimiter: string(mb.Delim),
			Flags:     flags,
		})
	}
	return result, nil
}

func (ic *IMAPClient) ListMessages(mailbox string, page, limit uint32) ([]MessageHeader, uint32, error) {
	status, err := ic.client.Status(mailbox, &imap.StatusOptions{NumMessages: true}).Wait()
	if err != nil {
		return nil, 0, fmt.Errorf("status mailbox %q: %w", mailbox, err)
	}
	if status.NumMessages == nil || *status.NumMessages == 0 {
		return []MessageHeader{}, 0, nil
	}
	total := *status.NumMessages

	selectData, err := ic.client.Select(mailbox, &imap.SelectOptions{ReadOnly: true}).Wait()
	if err != nil {
		return nil, 0, fmt.Errorf("select mailbox %q: %w", mailbox, err)
	}
	totalCount := selectData.NumMessages
	if totalCount == 0 {
		return []MessageHeader{}, total, nil
	}

	offset := page * limit

	if offset >= totalCount {
		return []MessageHeader{}, total, nil
	}

	end := totalCount - offset
	var start uint32 = 1
	if end > limit {
		start = end - limit + 1
	}

	var seqSet imap.SeqSet
	seqSet.AddRange(start, end)
	fetchOptions := &imap.FetchOptions{
		Flags:    true,
		Envelope: true,
		UID:      true,
	}

	msgs, err := ic.client.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		return nil, 0, fmt.Errorf("fetch headers: %w", err)
	}

	headers := make([]MessageHeader, 0, len(msgs))
	for _, msg := range msgs {
		seen := false
		for _, f := range msg.Flags {
			if f == imap.FlagSeen {
				seen = true
				break
			}
		}

		from := ""
		if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
			addr := msg.Envelope.From[0]
			if addr.Name != "" {
				from = addr.Name
			} else {
				from = addr.Mailbox + "@" + addr.Host
			}
		}

		subject := ""
		if msg.Envelope != nil {
			subject = msg.Envelope.Subject
		}

		date := msg.Envelope.Date

		headers = append(headers, MessageHeader{
			UID:     uint32(msg.UID),
			Subject: subject,
			From:    from,
			Date:    date,
			Seen:    seen,
		})
	}

	for i, j := 0, len(headers)-1; i < j; i, j = i+1, j-1 {
		headers[i], headers[j] = headers[j], headers[i]
	}

	return headers, total, nil
}

func (ic *IMAPClient) DeleteMessage(mailbox string, uid uint32) error {
	if _, err := ic.client.Select(mailbox, nil).Wait(); err != nil {
		return fmt.Errorf("select mailbox: %w", err)
	}

	uidSet := imap.UIDSetNum(imap.UID(uid))

	storeFlags := &imap.StoreFlags{
		Op:     imap.StoreFlagsAdd,
		Flags:  []imap.Flag{imap.FlagDeleted},
		Silent: true,
	}
	if err := ic.client.Store(uidSet, storeFlags, nil).Close(); err != nil {
		return fmt.Errorf("flag deleted: %w", err)
	}

	if err := ic.client.Expunge().Close(); err != nil {
		return fmt.Errorf("expunge: %w", err)
	}

	return nil
}

func (ic *IMAPClient) MarkMessage(mailbox string, uid uint32, seen bool) error {
	if _, err := ic.client.Select(mailbox, nil).Wait(); err != nil {
		return fmt.Errorf("select mailbox: %w", err)
	}

	uidSet := imap.UIDSetNum(imap.UID(uid))
	op := imap.StoreFlagsAdd
	if !seen {
		op = imap.StoreFlagsDel
	}

	storeFlags := &imap.StoreFlags{
		Op:     op,
		Flags:  []imap.Flag{imap.FlagSeen},
		Silent: true,
	}
	if err := ic.client.Store(uidSet, storeFlags, nil).Close(); err != nil {
		return fmt.Errorf("mark seen=%v: %w", seen, err)
	}
	return nil
}

func (ic *IMAPClient) SearchMessages(mailbox, query string) ([]MessageHeader, error) {
	if _, err := ic.client.Select(mailbox, &imap.SelectOptions{ReadOnly: true}).Wait(); err != nil {
		return nil, fmt.Errorf("select mailbox: %w", err)
	}

	criteria := &imap.SearchCriteria{
		Or: [][2]imap.SearchCriteria{
			{
				{Text: []string{query}},
				{Header: []imap.SearchCriteriaHeaderField{{Key: "From", Value: query}}},
			},
		},
	}

	searchData, err := ic.client.Search(criteria, &imap.SearchOptions{ReturnAll: true}).Wait()
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	if len(searchData.AllSeqNums()) == 0 {
		return []MessageHeader{}, nil
	}

	fetchOptions := &imap.FetchOptions{
		Flags:    true,
		Envelope: true,
		UID:      true,
	}

	msgs, err := ic.client.Fetch(searchData.All, fetchOptions).Collect()
	if err != nil {
		return nil, fmt.Errorf("fetch search results: %w", err)
	}

	headers := make([]MessageHeader, 0, len(msgs))
	for _, msg := range msgs {
		seen := false
		for _, f := range msg.Flags {
			if f == imap.FlagSeen {
				seen = true
				break
			}
		}
		from := ""
		if msg.Envelope != nil && len(msg.Envelope.From) > 0 {
			a := msg.Envelope.From[0]
			if a.Name != "" {
				from = a.Name
			} else {
				from = a.Mailbox + "@" + a.Host
			}
		}
		subject := ""
		if msg.Envelope != nil {
			subject = msg.Envelope.Subject
		}
		headers = append(headers, MessageHeader{
			UID:     uint32(msg.UID),
			Subject: subject,
			From:    from,
			Date:    msg.Envelope.Date,
			Seen:    seen,
		})
	}

	for i, j := 0, len(headers)-1; i < j; i, j = i+1, j-1 {
		headers[i], headers[j] = headers[j], headers[i]
	}

	return headers, nil
}

func (ic *IMAPClient) UnseenCount(mailbox string) (uint32, error) {
	status, err := ic.client.Status(mailbox, &imap.StatusOptions{NumUnseen: true}).Wait()
	if err != nil {
		return 0, fmt.Errorf("status: %w", err)
	}
	if status.NumUnseen == nil {
		return 0, nil
	}
	return *status.NumUnseen, nil
}

func (ic *IMAPClient) FetchMessage(mailbox string, uid uint32) (*Message, error) {
	if _, err := ic.client.Select(mailbox, nil).Wait(); err != nil {
		return nil, fmt.Errorf("select mailbox: %w", err)
	}

	uidSet := imap.UIDSetNum(imap.UID(uid))
	fetchOptions := &imap.FetchOptions{
		Flags:       true,
		Envelope:    true,
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{{}},
	}

	msgs, err := ic.client.Fetch(uidSet, fetchOptions).Collect()
	if err != nil || len(msgs) == 0 {
		return nil, fmt.Errorf("fetch message uid=%d: %w", uid, err)
	}

	raw := msgs[0]
	msg := &Message{}
	msg.UID = uid

	if raw.Envelope != nil {
		msg.Subject = raw.Envelope.Subject
		msg.Date = raw.Envelope.Date

		if len(raw.Envelope.From) > 0 {
			a := raw.Envelope.From[0]
			msg.From = a.Name
			if msg.From == "" {
				msg.From = a.Mailbox + "@" + a.Host
			}
		}

		if len(raw.Envelope.ReplyTo) > 0 {
			a := raw.Envelope.ReplyTo[0]
			msg.ReplyTo = a.Mailbox + "@" + a.Host
		} else {
			msg.ReplyTo = msg.From
		}

		for _, a := range raw.Envelope.To {
			msg.To = append(msg.To, a.Mailbox+"@"+a.Host)
		}

		for _, a := range raw.Envelope.Cc {
			msg.CC = append(msg.CC, a.Mailbox+"@"+a.Host)
		}
	}

	for _, sectionBytes := range raw.BodySection {
		if err := parseMIME(msg, sectionBytes); err != nil {
			msg.TextBody = strings.TrimSpace(string(sectionBytes))
		}
		break
	}

	storeFlags := &imap.StoreFlags{
		Op:     imap.StoreFlagsAdd,
		Flags:  []imap.Flag{imap.FlagSeen},
		Silent: true,
	}
	_ = ic.client.Store(uidSet, storeFlags, nil).Close()

	return msg, nil
}

func parseMIME(msg *Message, raw []byte) error {
	m, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return err
	}

	ct := m.Header.Get("Content-Type")
	if ct == "" {
		ct = "text/plain"
	}

	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		mediaType = "text/plain"
		params = map[string]string{}
	}

	return walkPart(msg, m.Body, mediaType, params, m.Header.Get("Content-Transfer-Encoding"), "")
}

func walkPart(msg *Message, r io.Reader, mediaType string, params map[string]string, cte string, filename string) error {
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary, ok := params["boundary"]
		if !ok {
			return fmt.Errorf("multipart missing boundary")
		}
		mr := multipart.NewReader(r, boundary)
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			partCT := part.Header.Get("Content-Type")
			if partCT == "" {
				partCT = "text/plain"
			}
			partMediaType, partParams, err := mime.ParseMediaType(partCT)
			if err != nil {
				partMediaType = "text/plain"
				partParams = map[string]string{}
			}

			partCTE := part.Header.Get("Content-Transfer-Encoding")

			partFilename := ""
			if cd := part.Header.Get("Content-Disposition"); cd != "" {
				_, cdParams, err := mime.ParseMediaType(cd)
				if err == nil {
					partFilename = cdParams["filename"]
				}
			}
			if partFilename == "" {
				partFilename = partParams["name"]
			}

			if err := walkPart(msg, part, partMediaType, partParams, partCTE, partFilename); err != nil {
				return err
			}
		}
		return nil
	}

	var decoded io.Reader = r
	switch strings.ToLower(strings.TrimSpace(cte)) {
	case "quoted-printable":
		decoded = quotedprintable.NewReader(r)
	case "base64":
		decoded = newBase64Reader(r)
	}

	data, err := io.ReadAll(decoded)
	if err != nil {
		return err
	}

	switch {
	case filename != "":
		msg.Attachments = append(msg.Attachments, Attachment{
			Filename:    filename,
			ContentType: mediaType,
			Size:        len(data),
			Data:        data,
		})
	case mediaType == "text/html" && msg.HTMLBody == "":
		msg.HTMLBody = string(data)
	default:
		if msg.TextBody == "" {
			msg.TextBody = strings.TrimSpace(string(data))
		}
	}

	return nil
}
