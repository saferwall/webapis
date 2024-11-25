// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package mailer

// Attachment interface
type Attachment interface {
	Base64Data() string
	Name() string
	MimeType() string
}

// Mailer interface
type Mailer interface {
	Send(body, subject, from, to string, attachments []Attachment) error
}
