package handler

import (
	"net/http"
	"strconv"

	"nagae/internal/mail"
	"nagae/internal/session"
	inboxview "nagae/views/inbox"
)

const pageSize = 50

func InboxGet(w http.ResponseWriter, r *http.Request) {
	creds := session.GetCredentials(r)

	mailbox := r.URL.Query().Get("mailbox")
	if mailbox == "" {
		mailbox = "INBOX"
	}

	page := uint32(0)
	if p, err := strconv.ParseUint(r.URL.Query().Get("page"), 10, 32); err == nil {
		page = uint32(p)
	}

	flash := ""
	if r.URL.Query().Get("deleted") == "1" {
		flash = "message deleted."
	}
	if r.URL.Query().Get("sent") == "1" {
		flash = "message sent."
	}

	client, err := mail.NewIMAPClient(creds)
	if err != nil {
		http.Error(w, "could not connect to mail server", http.StatusBadGateway)
		return
	}
	defer client.Close()

	mailboxes, err := client.ListMailboxes()
	if err != nil {
		http.Error(w, "could not list mailboxes", http.StatusInternalServerError)
		return
	}

	messages, total, err := client.ListMessages(mailbox, page, pageSize)
	if err != nil {
		http.Error(w, "could not fetch messages", http.StatusInternalServerError)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	w.Header().Set("Cache-Control", "no-store")
	inboxview.Inbox(creds.Username, mailbox, mailboxes, messages, flash, inboxview.Pagination{
		Page:       page,
		TotalPages: totalPages,
		Total:      total,
		Mailbox:    mailbox,
	}).Render(r.Context(), w)
}

func MessageGet(w http.ResponseWriter, r *http.Request) {
	creds := session.GetCredentials(r)

	mailbox := r.URL.Query().Get("mailbox")
	if mailbox == "" {
		mailbox = "INBOX"
	}

	uidStr := r.URL.Query().Get("uid")
	uid64, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid uid", http.StatusBadRequest)
		return
	}

	client, err := mail.NewIMAPClient(creds)
	if err != nil {
		http.Error(w, "could not connect to mail server", http.StatusBadGateway)
		return
	}
	defer client.Close()

	mailboxes, _ := client.ListMailboxes()

	msg, err := client.FetchMessage(mailbox, uint32(uid64))
	if err != nil {
		http.Error(w, "could not fetch message", http.StatusInternalServerError)
		return
	}

	inboxview.Message(creds.Username, mailbox, mailboxes, msg).Render(r.Context(), w)
}
