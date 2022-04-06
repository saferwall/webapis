package file

import (
	"reflect"
	"strings"
)

// IsFilterAllowed check if we are allowed to filter GET with fields
func IsFilterAllowed(allowed []string, filters []string) bool {
	for _, filter := range filters {
		if !IsStringInSlice(filter, allowed) {
			return false
		}
	}
	return true
}

// GetStructFields retrieve json struct fields names
func GetStructFields(i interface{}) []string {

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

// IsStringInSlice check if a string exist in a list of strings
func IsStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}
