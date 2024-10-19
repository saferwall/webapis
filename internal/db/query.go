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
	ActionFollow n1qlQuery = iota
	ActionLike
	ActionSubmit
	ActionUnfollow
	ActionUnlike
	AnoUserActivities
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
	MetaUI
	UserActivities
	UserComments
	UserFollowers
	UserFollowing
	UserLikes
	UserSubmissions
)

var fileQueryMap = map[string]n1qlQuery{
	"action-follow.sql":             ActionFollow,
	"action-like.sql":               ActionLike,
	"action-submit.sql":             ActionSubmit,
	"action-unfollow.sql":           ActionUnfollow,
	"action-unlike.sql":             ActionUnlike,
	"ano-user-activities.sql":       AnoUserActivities,
	"ano-user-comments.sql":         AnoUserComments,
	"ano-user-followers.sql":        AnoUserFollowers,
	"ano-user-following.sql":        AnoUserFollowing,
	"ano-user-likes.sql":            AnoUserLikes,
	"ano-user-submissions.sql":      AnoUserSubmissions,
	"behavior-report.sql":           BehaviorReport,
	"count-ano-user-activities.sql": CountAnoUserActivities,
	"count-strings.sql":             CountStrings,
	"count-user-activities.sql":     CountUserActivities,
	"delete-activity.sql":           DeleteActivity,
	"file-comments.sql":             FileComments,
	"file-strings.sql":              FileStrings,
	"file-summary.sql":              FileSummary,
	"get-all-doc-type.sql":          GetAllDocType,
	"meta-ui.sql":                   MetaUI,
	"user-activities.sql":           UserActivities,
	"user-comments.sql":             UserComments,
	"user-followers.sql":            UserFollowers,
	"user-following.sql":            UserFollowing,
	"user-likes.sql":                UserLikes,
	"user-submissions.sql":          UserSubmissions,
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
