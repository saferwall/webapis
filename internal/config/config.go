// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

// Package consumer implements the NSQ worker logic.
package config

import (
	"os"

	"github.com/spf13/viper"
)

// DatabaseCfg represents the database config.
type DatabaseCfg struct {
	// the data source name (DSN) for connecting to the database.
	Server string `mapstructure:"server"`
	// Username used to access the db.
	Username string `mapstructure:"username"`
	// Password used to access the db.
	Password string `mapstructure:"password"`
	// Name of the couchbase bucket.
	BucketName string `mapstructure:"bucket_name"`
}

// NsqCfg represents NSQ config.
type NsqCfg struct {
	// the data source name (DSN) for connecting to the nsqd.
	Address string `mapstructure:"address"`
}

// UICfg represents frontend config.
type UICfg struct {
	// the data source name (DSN) for connecting to the frontend.
	Address string `mapstructure:"address"`
}

// StorageCfg represents the object storage config.
type StorageCfg struct {
	Endpoint  string `mapstructure:"endpoint"`
	SecKey    string `mapstructure:"seckey"`
	AccessKey string `mapstructure:"accesskey"`
	Spacename string `mapstructure:"spacename"`
	Ssl       bool   `mapstructure:"ssl"`
}

// Config represents our application config.
type Config struct {
	// The IP:Port. Defaults to 8080.
	Address string `mapstructure:"address"`
	// Log level. Defaults to info.
	LogLevel string `mapstructure:"log_level"`
	// The data source name (DSN) for th frontend.
	FrontendAddress string `mapstructure:"frontend_address"`
	// Maximum file size to allow for samples.
	MaxFileSize int64 `mapstructure:"max_file_size"`
	// Maximum avatar size to allow for user profile picture.
	MaxAvatarSize int64 `mapstructure:"max_avatar_file_size"`
	// Database configuration.
	DB DatabaseCfg `mapstructure:"db"`
	// NSQ configuration.
	Nsq NsqCfg `mapstructure:"nsq"`
	// Frontend Configuration.
	UI UICfg `mapstructure:"ui"`
	// Object storage configuration.
	ObjStorage StorageCfg `mapstructure:"storage"`
}

// Load returns an application configuration which is populated
// from the given configuration file.
func Load(path string) (*Config, error) {

	// Create a new config.
	c := Config{}

	// Adding our TOML config file.
	viper.AddConfigPath(path)

	// Load the config type depending on env variable.
	var name string
	env := os.Getenv("SFW_WEB_APP")
	switch env {
	case "local":
		name = "local"
	case "dev":
		name = "dev"
	case "prod":
		name = "prod"
	default:
		name = "local"
	}

	// Set the config name to choose from the config path
	// Extension not needed.
	viper.SetConfigName(name)

	// Load the configuration from disk.
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	// Unmarshals the config into a Struct.
	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, err
}
