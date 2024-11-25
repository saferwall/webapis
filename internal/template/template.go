// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package tpl

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type EmailTemplate int

const (
	ConfirmAccount = iota
	ResetPassword
	EmailUpdate
)

var emailTplMap = map[string]EmailTemplate{
	"account-confirmation": ConfirmAccount,
	"password-reset":       ResetPassword,
	"email-update":         EmailUpdate,
}

var (
	regExtractB64 = regexp.MustCompile(`data:(image\/\w{2,5});base64,([^"]*)`)
)

type Service struct {
	EmailRequestTemplate map[EmailTemplate]EmailRequest
}

type EmailAttachement struct {
	B64Data  string
	FileName string
	Type     string
}

// EmailRequest struct
type EmailRequest struct {
	From       string
	Subject    string
	tpl        *template.Template
	InlineImgs []EmailAttachement
}

func (ea EmailAttachement) Base64Data() string {
	return ea.B64Data

}
func (ea EmailAttachement) Name() string {
	return ea.FileName

}
func (ea EmailAttachement) MimeType() string {
	return ea.Type

}

// walk returns list of files in directory.
func walk(dir string) ([]string, error) {

	fileList := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() && strings.HasSuffix(path, ".html") {
			fileList = append(fileList, path)
		}
		return nil
	})

	return fileList, err
}

// New iterate over the list of html templates and parses them.
func New(filePath string) (Service, error) {
	htmlTemplates, err := walk(filePath)
	if err != nil {
		return Service{}, err
	}

	templates := make(map[EmailTemplate]EmailRequest)
	for _, f := range htmlTemplates {

		er := EmailRequest{
			From: "noreply@saferwall.com",
		}
		// Extract Base64 inline images
		data, err := os.ReadFile(f)
		if err != nil {
			return Service{}, nil
		}

		i := 0
		match := regExtractB64.ReplaceAllStringFunc(string(data), func(s string) string {
			matches := regExtractB64.FindStringSubmatch(s)
			emailAttach := EmailAttachement{}
			emailAttach.Type = matches[1]
			emailAttach.B64Data = matches[2]
			emailAttach.FileName = "filename" + strconv.Itoa(i)
			er.InlineImgs = append(er.InlineImgs, emailAttach)
			i++
			return "cid:" + emailAttach.FileName
		})

		dir := filepath.Dir(f)
		name := filepath.Base(dir)
		key := emailTplMap[name]

		t, err := template.New(name).Parse(match)
		if err != nil {
			return Service{}, err
		}

		er.tpl = t

		switch name {
		case "account-confirmation":
			er.Subject = "saferwall - confirm account"
		case "password-reset":
			er.Subject = "saferwall - reset password"
		case "email-update":
			er.Subject = "saferwall - confirm new email"
		}
		templates[key] = er
	}

	return Service{templates}, nil
}

func (er EmailRequest) Execute(templateData interface{}, wr io.Writer) error {
	buf := new(bytes.Buffer)
	if err := er.tpl.Execute(buf, templateData); err != nil {
		return err
	}
	_, err := wr.Write(buf.Bytes())
	return err
}
