// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package mailer

import (
	"time"

	"github.com/saferwall/saferwall-api/pkg/log"
	mail "github.com/xhit/go-simple-mail/v2"
)

type SMTPServer struct {
	server *mail.SMTPServer
}

type SMTPClient struct {
	client *mail.SMTPClient
	logger log.Logger
	quit   chan struct{}
}

// New creates a new SMTP client using the default configuration.
func New(host string, port int, username, password string) SMTPServer {
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

	// Variable to keep alive connection
	server.KeepAlive = true

	// Timeout for connect to SMTP Server
	server.ConnectTimeout = 10 * time.Second

	// Timeout for send the data and wait respond
	server.SendTimeout = 10 * time.Second

	return SMTPServer{server}
}

func (s SMTPServer) Connect(logger log.Logger) (SMTPClient, error) {
	c, err := s.server.Connect()
	if err != nil {
		return SMTPClient{}, err
	}

	// NOOP command, optional, used for avoid timeout when KeepAlive is true and
	// you aren't sending mails. Execute this command each 30 seconds is ideal
	// for persistent connection
	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				err = c.Noop()
				if err != nil {
					logger.Error("failed to noop smtp, reason: %v", err)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return SMTPClient{c, logger, quit}, nil
}

func (c SMTPClient) Send(htmlBody, subject, from, to string) error {

	if c.client == nil {
		return nil
	}

	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(subject)
	email.SetBody(mail.TextHTML, htmlBody)

	// always check error after send
	if email.Error != nil {
		return email.Error
	}

	// Call Send and pass the client
	if err := email.Send(c.client); err != nil {
		c.logger.Errorf("email send failed: %v", err)
	}

	return nil
}
