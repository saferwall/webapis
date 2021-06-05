package main

import (
	"github.com/saferwall/saferwall-api/pkg/log"
)


// Version indicates the current version of the application.
var Version = "1.0.0"


func main() {

	// create root logger tagged with server version
	logger := log.New().With(nil, "version", Version)

	logger.Info("Hello DDD")

}