package file

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
)

// isFilterAllowed check if we are allowed to filter GET with fields
func isFilterAllowed(allowed []string, filters []string) bool {
	for _, filter := range filters {
		if !isStringInSlice(filter, allowed) {
			return false
		}
	}
	return true
}

// getStructFields retrieve json struct fields names
func getStructFields(i interface{}) []string {

	val := reflect.ValueOf(i)
	var temp string

	var listFields []string
	for i := 0; i < val.Type().NumField(); i++ {
		temp = val.Type().Field(i).Tag.Get("json")
		temp = strings.Replace(temp, ",omitempty", "", -1)
		listFields = append(listFields, temp)
	}

	return listFields
}

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

// hash calculates the sha256 hash over a stream of bytes.
func hash(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
