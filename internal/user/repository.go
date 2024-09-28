// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package user

import (
	"context"
	"encoding/json"
	"strings"

	dbcontext "github.com/saferwall/saferwall-api/internal/db"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/pkg/log"
)

// Repository encapsulates the logic to access users from the data source.
type Repository interface {
	// Get returns the user with the specified user ID.
	Get(ctx context.Context, id string) (entity.User, error)
	// Count returns the number of users.
	Count(ctx context.Context) (int, error)
	// Query returns the list of users with the given offset and limit.
	Query(ctx context.Context, offset, limit int) ([]entity.User, error)
	// Create saves a new user in the storage.
	Create(ctx context.Context, User entity.User) error
	// Update updates the user with given ID in the storage.
	Update(ctx context.Context, User entity.User) error
	// Patch patches a sub entry in the user with given ID in the storage.
	Patch(ctx context.Context, key, path string, val interface{}) error
	// Delete removes the user with given ID from the storage.
	Delete(ctx context.Context, id string) error
	// Exists checks if a user exists with a given ID.
	Exists(ctx context.Context, id string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	Likes(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Followers(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Following(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Submissions(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Comments(ctx context.Context, id string, offset, limit int) (
		[]interface{}, error)
	Activities(ctx context.Context, id string, offset, limit int) ([]interface{}, error)
	CountActivities(ctx context.Context) (int, error)
	Like(ctx context.Context, id string, userLike entity.UserLike) error
	Unlike(ctx context.Context, id, sha256 string) error
}

// repository persists users in database.
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new user repository.
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the user with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.User, error) {
	var user entity.User
	key := strings.ToLower(id)
	err := r.db.Get(ctx, key, &user)
	return user, err
}

// Create saves a new user record in the database.
// It returns the ID of the newly inserted user record.
func (r repository) Create(ctx context.Context, user entity.User) error {
	key := user.ID()
	return r.db.Create(ctx, key, &user)
}

// Update saves the changes to a user in the database.
func (r repository) Update(ctx context.Context, user entity.User) error {
	key := user.ID()
	return r.db.Update(ctx, key, &user)
}

// Patch performs a sub doc update to a user in the database.
func (r repository) Patch(ctx context.Context, key, path string,
	val interface{}) error {
	return r.db.Patch(ctx, key, path, val)
}

// Delete deletes a user with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	key := strings.ToLower(id)
	return r.db.Delete(ctx, key)
}

// Exists checks if a document exists for the given id.
func (r repository) Exists(ctx context.Context, id string) (bool, error) {
	docExists := false
	key := strings.ToLower(id)
	err := r.db.Exists(ctx, key, &docExists)
	return docExists, err
}

// Count returns the number of the user records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int

	params := make(map[string]interface{}, 1)
	params["docType"] = "user"

	statement :=
		"SELECT RAW COUNT(*) AS count FROM `" + r.db.Bucket.Name() + "` " +
			"WHERE `type`=$docType"

	err := r.db.Count(ctx, statement, params, &count)
	return count, err
}

// EmailExists states if the email address exists in the database.
func (r repository) EmailExists(ctx context.Context, email string) (bool, error) {
	var res interface{}

	params := make(map[string]interface{}, 1)
	params["docType"] = "user"
	params["email"] = strings.ToLower(email)

	statement :=
		"SELECT true as docPresent FROM `" + r.db.Bucket.Name() + "` " +
			"WHERE `type`=$docType AND email=$email"

	if err := r.db.Query(ctx, statement, params, &res); err != nil {
		return false, err
	}

	present := res.([]interface{})
	return len(present) > 0, nil
}

// GetByEmail states if the email address exists in the database.
func (r repository) GetByEmail(ctx context.Context, email string) (
	entity.User, error) {
	var res interface{}

	params := make(map[string]interface{}, 1)
	params["docType"] = "user"
	params["email"] = strings.ToLower(email)

	statement :=
		"SELECT " + r.db.Bucket.Name() + ".*  FROM `" + r.db.Bucket.Name() +
			"` WHERE `type`=$docType AND email=$email"

	if err := r.db.Query(ctx, statement, params, &res); err != nil {
		return entity.User{}, nil
	}

	if res.([]interface{}) == nil {
		return entity.User{}, errUserNotFound
	}

	user := entity.User{}
	b, _ := json.Marshal(res.([]interface{})[0])
	_ = json.Unmarshal(b, &user)
	return user, nil
}

// Query retrieves the user records with the specified offset and limit
// from the database.
func (r repository) Query(ctx context.Context, offset, limit int) (
	[]entity.User, error) {
	var res interface{}

	params := make(map[string]interface{}, 1)
	params["docType"] = "user"
	params["offset"] = offset
	params["limit"] = limit

	statement := r.db.N1QLQuery[dbcontext.GetAllDocType]
	err := r.db.Query(ctx, statement, params, &res)
	if err != nil {
		return []entity.User{}, err
	}
	users := []entity.User{}
	for _, u := range res.([]interface{}) {
		user := entity.User{}
		b, _ := json.Marshal(u)
		_ = json.Unmarshal(b, &user)
		user.Email = ""
		user.Password = ""
		users = append(users, user)
	}
	return users, nil
}

func (r repository) Activities(ctx context.Context, id string, offset,
	limit int) ([]interface{}, error) {

	var activities interface{}
	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit

	if id == "" {
		// For an anonymous user.
		err := r.db.Query(ctx, r.db.N1QLQuery[dbcontext.AnoUserActivities], params, &activities)
		if err != nil {
			return nil, err
		}
	} else {
		// For a logged-in user.
		params["user"] = id
		err := r.db.Query(ctx, r.db.N1QLQuery[dbcontext.UserActivities], params, &activities)
		if err != nil {
			return nil, err
		}
	}

	return activities.([]interface{}), nil

}

func (r repository) Likes(ctx context.Context, id string, offset,
	limit int) ([]interface{}, error) {

	var results interface{}
	var currentUser, query string
	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit
	params["user"] = strings.ToLower(id)

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		currentUser = user.ID()
	}

	if currentUser == "" {
		// For an anonymous user.
		query = r.db.N1QLQuery[dbcontext.AnoUserLikes]
	} else {
		// For a logged-in user.
		params["loggedInUser"] = currentUser
		query = r.db.N1QLQuery[dbcontext.UserLikes]
	}

	err := r.db.Query(ctx, query, params, &results)
	if err != nil {
		return nil, err
	}
	if len(results.([]interface{})) == 0 {
		return []interface{}{}, nil
	}
	return results.([]interface{}), nil
}

func (r repository) Followers(ctx context.Context, id string, offset,
	limit int) ([]interface{}, error) {

	var results interface{}
	var currentUser, query string
	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit
	params["user"] = strings.ToLower(id)

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		currentUser = user.ID()
	}

	if currentUser == "" {
		// For an anonymous user.
		query = r.db.N1QLQuery[dbcontext.AnoUserFollowers]
	} else {
		// For a logged-in user.
		params["loggedInUser"] = currentUser
		query = r.db.N1QLQuery[dbcontext.UserFollowers]
	}

	err := r.db.Query(ctx, query, params, &results)
	if err != nil {
		return nil, err
	}
	if len(results.([]interface{})) == 0 {
		return []interface{}{}, nil
	}
	return results.([]interface{}), nil
}

func (r repository) Following(ctx context.Context, id string, offset,
	limit int) ([]interface{}, error) {

	var results interface{}
	var currentUser, query string
	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit
	params["user"] = strings.ToLower(id)

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		currentUser = user.ID()
	}

	if currentUser == "" {
		// For an anonymous user.
		query = r.db.N1QLQuery[dbcontext.AnoUserFollowing]
	} else {
		// For a logged-in user.
		params["loggedInUser"] = currentUser
		query = r.db.N1QLQuery[dbcontext.UserFollowing]
	}

	err := r.db.Query(ctx, query, params, &results)
	if err != nil {
		return nil, err
	}
	if len(results.([]interface{})) == 0 {
		return []interface{}{}, nil
	}
	return results.([]interface{}), nil
}

func (r repository) Submissions(ctx context.Context, id string, offset,
	limit int) ([]interface{}, error) {
	var results interface{}
	var currentUser, query string
	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit
	params["user"] = strings.ToLower(id)

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		currentUser = user.ID()
	}

	if currentUser == "" {
		// For an anonymous user.
		query = r.db.N1QLQuery[dbcontext.AnoUserSubmissions]
	} else {
		// For a logged-in user.
		params["loggedInUser"] = currentUser
		query = r.db.N1QLQuery[dbcontext.UserSubmissions]
	}

	err := r.db.Query(ctx, query, params, &results)
	if err != nil {
		return nil, err
	}
	if len(results.([]interface{})) == 0 {
		return []interface{}{}, nil
	}
	return results.([]interface{}), nil
}

func (r repository) Comments(ctx context.Context, id string, offset,
	limit int) ([]interface{}, error) {

	var results interface{}
	var currentUser, query string
	params := make(map[string]interface{}, 1)
	params["offset"] = offset
	params["limit"] = limit
	params["user"] = strings.ToLower(id)

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		currentUser = user.ID()
	}

	if currentUser == "" {
		// For an anonymous user.
		query = r.db.N1QLQuery[dbcontext.AnoUserComments]
	} else {
		// For a logged-in user.
		params["loggedInUser"] = currentUser
		query = r.db.N1QLQuery[dbcontext.UserComments]
	}

	err := r.db.Query(ctx, query, params, &results)
	if err != nil {
		return nil, err
	}
	if len(results.([]interface{})) == 0 {
		return []interface{}{}, nil
	}
	return results.([]interface{}), nil
}

func (r repository) CountActivities(ctx context.Context) (int, error) {

	var count int
	var query string
	var currentUser string
	params := make(map[string]interface{}, 1)

	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		currentUser = user.ID()
	}

	if currentUser == "" {
		// For an anonymous user.
		query = r.db.N1QLQuery[dbcontext.CountAnoUserActivities]
	} else {
		// For a logged-in user.
		params["user"] = currentUser
		query = r.db.N1QLQuery[dbcontext.CountUserActivities]

	}
	err := r.db.Count(ctx, query, params, &count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r repository) Like(ctx context.Context, id string, userLike entity.UserLike) error {

	var results interface{}

	params := make(map[string]interface{}, 1)
	params["sha256"] = userLike.SHA256
	params["ts"] = userLike.Timestamp
	params["user"] = id

	query := r.db.N1QLQuery[dbcontext.ActionLike]
	return r.db.Query(ctx, query, params, &results)
}

func (r repository) Unlike(ctx context.Context, id, sha256 string) error {

	var results interface{}

	params := make(map[string]interface{}, 1)
	params["sha256"] = sha256
	params["user"] = id

	query := r.db.N1QLQuery[dbcontext.ActionUnlike]
	return r.db.Query(ctx, query, params, &results)

}
