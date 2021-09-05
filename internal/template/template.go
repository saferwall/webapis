// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package tpl

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type EmailTemplate int

const (
	ConfirmAccount = iota
	ResetPassword
)

var emailTplMap = map[string]EmailTemplate{
	"account-confirmation": ConfirmAccount,
	"password-reset":       ResetPassword,
}

type Service struct {
	EmailRequestTemplate map[EmailTemplate]EmailRequest
}

// EmailRequest struct
type EmailRequest struct {
	From    string
	Subject string
	tpl     *template.Template
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
		t, err := template.ParseFiles(f)
		if err != nil {
			return Service{}, err
		}

		dir := filepath.Dir(f)
		name := filepath.Base(dir)
		key := emailTplMap[name]

		er := EmailRequest{
			From: "noreply@saferwall.com",
			tpl: t,
		}

		switch name {
		case "account-confirmation":
			er.Subject = "saferwall - confirm account"
		case "password-reset":
			er.Subject = "saferwall - reset password"
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
	wr.Write(buf.Bytes())
	return nil
}
