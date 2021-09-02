// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package archive

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeka/zip"
)

type Archiver struct {
	enc zip.EncryptionMethod
}

// New initializes security service.
func New(enc zip.EncryptionMethod) Archiver {
	return Archiver{enc: enc}
}

// Archive filename binary data to zip using a password.
func (s Archiver) Archive(zipFilePath, password string, r io.Reader) error {

	fzip, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}

	zipw := zip.NewWriter(fzip)
	defer zipw.Close()

	filename := trimExt(filepath.Base(zipFilePath))
	w, err := zipw.Encrypt(filename, password, s.enc)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	zipw.Flush()
	return nil
}

// trimExt delete the extention from the file name.
func trimExt(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
