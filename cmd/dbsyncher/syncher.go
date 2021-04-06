package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/couchbase/gocb/v2"
	"github.com/jaswdr/faker"
	"github.com/saferwall/saferwall-api/app/handler/file"
	"github.com/saferwall/saferwall-api/app/handler/user"
	"golang.org/x/crypto/bcrypt"
)

const (
	prodHost    = "https://api.saferwall.com"
	dbServer    = "couchbase://localhost"
	dbUsername  = "Administrator"
	dbPassword  = "password"
	sha256regex = "^[a-f0-9]{64}$"
)

func httpGet(url string) ([]byte, error) {

	var body []byte
	resp, err := http.Get(url)
	if err != nil {
		return body, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("invalid HTTP response status code")
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	return body, err
}

func dbConnect(cbServer string, cbUsername string, cbPassword string) (
	*gocb.Collection, *gocb.Collection, error) {

	opts := gocb.ClusterOptions{
		Username: cbUsername,
		Password: cbPassword,
	}

	// Init our cluster.
	cluster, err := gocb.Connect(cbServer, opts)
	if err != nil {
		return nil, nil, err
	}

	// Get a bucket reference over users.
	usersBucket := cluster.Bucket("users")
	filesBucket := cluster.Bucket("files")

	// Get a collection reference.
	usersCollection := usersBucket.DefaultCollection()
	filesCollection := filesBucket.DefaultCollection()
	return usersCollection, filesCollection, nil
}

func HashAndSalt(pwd []byte) string {

	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return ""
	}

	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

func main() {

	// Connect to the local DB.
	usersCollection, filesCollection, err := dbConnect(
		dbServer, dbUsername, dbPassword)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Read the clipboard output.
	clipContent, err := clipboard.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read clipboard content: %v", err)
	}

	uniStr := strings.ReplaceAll(clipContent, "\r\n", "\n")
	list := strings.Split(uniStr, "\n")
	if len(list) == 0 {
		log.Fatalln("Empty list in clipboard")
	}

	matched, _ := regexp.MatchString(sha256regex, list[0])
	if matched {
		for _, sha256 := range list {
			url := prodHost + "/v1/files/" + sha256
			jsonBody, err := httpGet(url)
			if err != nil {
				log.Fatalf("Failed to read %s, err: %v", sha256, err)
				continue
			}

			// Unmarshal results.
			f := file.File{}
			err = json.Unmarshal(jsonBody, &f)
			if err != nil {
				log.Printf("Failed to unmarshal JSON %s, err: %v", sha256, err)
				continue
			}

			key := strings.ToLower(sha256)
			_, err = filesCollection.Upsert(key, f, &gocb.UpsertOptions{})
			if err != nil {
				log.Printf("Failed to json marshal object: %v", err)
				continue
			}
		}
	} else {
		for _, username := range list {
			log.Printf("Processing %s", username)

			url := prodHost + "/v1/users/" + username
			jsonBody, err := httpGet(url)
			if err != nil {
				log.Fatalf("Failed to read %s, err: %v", username, err)
				continue
			}

			// Unmarshal results.
			u := user.User{}
			err = json.Unmarshal(jsonBody, &u)
			if err != nil {
				log.Printf("Failed to unmarshal JSON %s, err: %v", username, err)
				continue
			}

			faker := faker.New()
			key := strings.ToLower(username)
			u.Email = faker.Internet().Email()
			u.Password = HashAndSalt([]byte(faker.Internet().Password()))
			_, err = usersCollection.Upsert(key, u, &gocb.UpsertOptions{})
			if err != nil {
				log.Printf("Failed to json marshal object: %v", err)
				continue
			}
		}
	}

}
