package main

import (
	"net/smtp"
	"strings"
)

type Mailer struct {
	Host       string
	Port       string
	Username   string
	Password   string
	From       string
	Recipients []string
}

func NewMailer(host, port, username, password, from string, recipients []string) *Mailer {
	return &Mailer{
		Host:       host,
		Port:       port,
		Username:   username,
		Password:   password,
		From:       from,
		Recipients: recipients,
	}
}

func (m *Mailer) SendMail(subject, body string) error {

	addr := m.Host + ":" + m.Port

	var auth smtp.Auth
	if m.Username != ""{
		auth = smtp.PlainAuth("", m.Username, m.Password, m.Host)
	}

	msg := []byte(
		"From: " + m.From + "\r\n" +
        "To: " + strings.Join(m.Recipients, ", ") + "\r\n" +
        "Subject: " + subject + "\r\n" +
        "Content-Type: text/plain\r\n" +
        "\r\n" +
        body,
	)

	return smtp.SendMail(addr, auth, m.From, m.Recipients, msg)
}
