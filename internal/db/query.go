// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package db

import (
	"os"
	"path/filepath"
	"strings"
)

type n1qlQuery int

const (
	AnoUserActivities n1qlQuery = iota
	AnoUserComments
	AnoUserFollowers
	AnoUserFollowing
	AnoUserLikes
	AnoUserSubmissions
	BehaviorReport
	CountAnoUserActivities
	CountStrings
	CountStringsWithSubstring
	CountUserActivities
	DeleteActivity
	FileComments
	FileStrings
	FileStringsWithSubstring
	FileSummary
	GetAllDocType
	UserActivities
	UserComments
	UserFollowers
	UserFollowing
	UserLikes
	UserSubmissions
	ActionLike
	ActionFollow
	ActionUnlike
	ActionUnfollow
	MetaUI
)

var fileQueryMap = map[string]n1qlQuery{
	"action-like.n1ql":               ActionLike,
	"action-follow.n1ql":             ActionFollow,
	"action-unlike.n1ql":             ActionUnlike,
	"action-unfollow.n1ql":           ActionUnfollow,
	"ano-user-activities.n1ql":       AnoUserActivities,
	"ano-user-comments.n1ql":         AnoUserComments,
	"ano-user-followers.n1ql":        AnoUserFollowers,
	"ano-user-following.n1ql":        AnoUserFollowing,
	"ano-user-likes.n1ql":            AnoUserLikes,
	"ano-user-submissions.n1ql":      AnoUserSubmissions,
	"behavior-report.n1ql":           BehaviorReport,
	"count-ano-user-activities.n1ql": CountAnoUserActivities,
	"count-strings.n1ql":             CountStrings,
	"count-user-activities.n1ql":     CountUserActivities,
	"delete-activity.n1ql":           DeleteActivity,
	"file-comments.n1ql":             FileComments,
	"file-strings.n1ql":              FileStrings,
	"file-summary.n1ql":              FileSummary,
	"get-all-doc-type.n1ql":          GetAllDocType,
	"user-activities.n1ql":           UserActivities,
	"user-comments.n1ql":             UserComments,
	"user-followers.n1ql":            UserFollowers,
	"user-following.n1ql":            UserFollowing,
	"user-likes.n1ql":                UserLikes,
	"user-submissions.n1ql":          UserSubmissions,
	"meta-ui.n1ql":                   MetaUI,
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
		c, err := os.ReadFile(f)
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
