// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package mailer

import (
	"os"
	"path/filepath"
	"text/template"
)

type emailTemplate int

const (
	ConfirmAccount = iota
	ResetPassword
)

var emailTplMap = map[string]emailTemplate{
	"confirm-account.html": ConfirmAccount,
	"reset-passwod.n1ql":   ResetPassword,
}

// walk returns list of files in directory.
func walk(dir string) ([]string, error) {

	fileList := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() {
			fileList = append(fileList, path)
		}
		return nil
	})

	return fileList, err
}

// ParseTemplates iterate over the list of html templates and parses them.
func ParseTemplates(filePath string) (
	map[emailTemplate]*template.Template, error) {

	htmlTemplates, err := walk(filePath)
	if err != nil {
		return nil, err
	}

	templates := make(map[emailTemplate]*template.Template)
	for _, f := range htmlTemplates {
		t, err := template.ParseFiles(f)
		if err != nil {
			return nil, err
		}

		name := filepath.Base(f)
		key := emailTplMap[name]
		templates[key] = t
	}

	return templates, nil
}
