// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package mailer

import (
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type SMTPMailer struct {
	server *mail.SMTPServer
}

// New creates a new SMTP mailer using the default configuration.
func New(host string, port int, username, password string) SMTPMailer {
	server := mail.NewSMTPClient()

	// SMTP Server
	server.Host = host
	server.Port = port
	server.Username = username
	server.Password = password
	server.Encryption = mail.EncryptionTLS

	// Since v2.3.0 you can specified authentication type:
	// - PLAIN (default)
	// - LOGIN
	// - CRAM-MD5
	server.Authentication = mail.AuthPlain

	// For thread safety, we do not want to keep alive connection.
	server.KeepAlive = false

	// Timeout for connect to SMTP Server
	server.ConnectTimeout = 10 * time.Second

	// Timeout for send the data and wait response.
	server.SendTimeout = 10 * time.Second

	return SMTPMailer{server: server}
}

func (s SMTPMailer) Send(htmlBody, subject, from, to string) error {
	if s.server == nil || s.server.Host == "" {
		return nil
	}

	client, err := s.server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(subject)
	email.SetBody(mail.TextHTML, htmlBody)

	// always check error after send
	if email.Error != nil {
		return email.Error
	}

	// Call Send and pass the client
	if err := email.Send(client); err != nil {
		return err
	}

	return nil
}
