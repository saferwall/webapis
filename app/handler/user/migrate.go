package user

import (
	"strings"

	"github.com/couchbase/gocb/v2"
	"github.com/saferwall/saferwall-api/app/common/db"
	log "github.com/sirupsen/logrus"
)

func getAllUsers() ([]User, error) {

	// Interfaces for handling streaming return values
	var row User
	var retValues []User

	// Select only demanded fields
	query := "SELECT users.* FROM `users`"

	// Execute Query
	results, err := db.Cluster.Query(query, &gocb.QueryOptions{})
	if err != nil {
		log.Errorf("Error executing n1ql query: %v", err)
		return retValues, nil
	}

	// Stream the values returned from the query into a typed array of structs
	for results.Next() {
		err := results.Row(&row)
		if err != nil {
			log.Errorf("results.Row() failed with: %v", err)
		}
		retValues = append(retValues, row)
		row = User{}
	}

	return retValues, nil
}

func UpdateProfileCounts() error {

	allUsers, err := getAllUsers()
	if err != nil {
		return err
	}

	for _, u := range allUsers {
		if len(u.Likes) == 0 && len(u.Comments) == 0 &&
			len(u.Followers) == 0 && len(u.Following) == 0 &&
			len(u.Submissions) == 0 {
			continue
		}
		log.Infof("Migrating %s", u.Username)
		u.LikesCount = len(u.Likes)
		u.CommentsCount = len(u.Comments)
		u.FollowersCount = len(u.Followers)
		u.FollowingCount = len(u.Following)
		u.SubmissionsCount = len(u.Submissions)
		u.Save()
	}

	return nil

}

func UpdateFollowingFollowersToLowerCase() error {

	allUsers, err := getAllUsers()
	if err != nil {
		return err
	}

	for _, u := range allUsers {
		log.Infof("Migrating %s", u.Username)

		var newFollowers  []string
		for _, f :=range u.Followers {
			newFollowers = append(newFollowers, strings.ToLower(f))
		}

		var newFollowing  []string
		for _, f :=range u.Following {
			newFollowing = append(newFollowing, strings.ToLower(f))
		}

		u.Followers = newFollowers
		u.Following = newFollowing
		if len(newFollowers) > 0 || len(newFollowing) > 0 {
			u.Save()
		}
	}

	return nil

}