// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package db

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type n1qlQuery int

const (
	UserActivities n1qlQuery = iota
	AnoUserActivities
	CountUserActivities
	CountAnoUserActivities
	UserLikes
	UserComments
	UserFollowing
	UserFollowers
	UserSubmissions
	AnoUserLikes
	AnoUserComments
	AnoUserFollowing
	AnoUserFollowers
	AnoUserSubmissions
	GetAllDocType
	DeleteActivity
	FileSummary
)

var fileQueryMap = map[string]n1qlQuery{
	"user-likes.n1ql":                UserLikes,
	"user-comments.n1ql":             UserComments,
	"user-following.n1ql":            UserFollowing,
	"user-followers.n1ql":            UserFollowers,
	"user-activities.n1ql":           UserActivities,
	"user-submissions.n1ql":          UserSubmissions,
	"ano-user-likes.n1ql":            AnoUserLikes,
	"ano-user-comments.n1ql":         AnoUserComments,
	"ano-user-following.n1ql":        AnoUserFollowing,
	"ano-user-followers.n1ql":        AnoUserFollowers,
	"ano-user-activities.n1ql":       AnoUserActivities,
	"ano-user-submissions.n1ql":      AnoUserSubmissions,
	"get-all-doc-type.n1ql":          GetAllDocType,
	"delete-activity.n1ql":           DeleteActivity,
	"count-user-activities.n1ql":     CountUserActivities,
	"count-ano-user-activities.n1ql": CountAnoUserActivities,
	"file-summary.n1ql":              FileSummary,
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

// PrepareQueries iterate over the list of n1ql files and load them to memory.
func (db *DB) PrepareQueries(filePath, bucketName string) error {

	n1qlFiles, err := walk(filePath)
	if err != nil {
		return err
	}

	db.N1QLQuery = make(map[n1qlQuery]string)
	for _, f := range n1qlFiles {
		c, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		// substitute the bucket name.
		query := string(c)
		query = strings.ReplaceAll(query, "bucket_name", bucketName)

		name := filepath.Base(f)
		key := fileQueryMap[name]
		db.N1QLQuery[key] = query
	}

	return nil
}
