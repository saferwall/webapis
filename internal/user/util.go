// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"bytes"
	"net/http"
)

// isStringInSlice check if a string exist in a list of strings
func isStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

// removeStringFromSlice removes a string item from a list of strings.
func removeStringFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// downloadURLContent reads a response from an URL content.
func downloadURLContent(url string) (*bytes.Buffer, error) {

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Perform the http post request.
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	// Read the response.
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return body, nil
}
