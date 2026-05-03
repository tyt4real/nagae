package handler

import "nagae/internal/config"

var mailServer *config.MailServer

func SetMailServer(ms *config.MailServer) {
	mailServer = ms
}
