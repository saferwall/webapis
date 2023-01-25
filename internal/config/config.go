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

// BrokerCfg represents the broker producer config.
type BrokerCfg struct {
	// the data source name (DSN) for connecting to the broker server.
	Address string `mapstructure:"address"`
	// Topic name to write to.
	Topic string `mapstructure:"topic"`
}

// UICfg represents frontend config.
type UICfg struct {
	// the data source name (DSN) for connecting to the frontend.
	Address string `mapstructure:"address"`
}

// AWSS3Cfg represents AWS S3 credentials.
type AWSS3Cfg struct {
	Region    string `mapstructure:"region"`
	SecretKey string `mapstructure:"secret_key"`
	AccessKey string `mapstructure:"access_key"`
}

// MinioCfg represents Minio credentials.
type MinioCfg struct {
	Endpoint  string `mapstructure:"endpoint"`
	SecretKey string `mapstructure:"secret_key"`
	AccessKey string `mapstructure:"access_key"`
	Region    string `mapstructure:"region"`
}

// LocalFsCfg represents local file system storage data.
type LocalFsCfg struct {
	RootDir string `mapstructure:"root_dir"`
}

// StorageCfg represents the object storage config.
type StorageCfg struct {
	// Deployment kind, possible values: aws, gcp, azure, local.
	DeploymentKind string `mapstructure:"deployment_kind"`
	// FileContainerName represents the name of the container for samples.
	FileContainerName string `mapstructure:"files_container_name"`
	// AvatarsContainerName represents the name of the container for avatars.
	AvatarsContainerName string `mapstructure:"avatars_container_name"`
	// S3 represents AWS S3 object storage connection details.
	S3 AWSS3Cfg `mapstructure:"s3"`
	// S3 represents MinIO object storage connection details.
	Minio MinioCfg `mapstructure:"minio"`
	// Local represents local file system config.
	Local LocalFsCfg `mapstructure:"local"`
}

type SMTPConfig struct {
	Server   string `mapstructure:"server"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// Config represents our application config.
type Config struct {
	// The IP:Port. Defaults to 8080.
	Address string `mapstructure:"address"`
	// Log level. Defaults to info.
	LogLevel string `mapstructure:"log_level"`
	// Disable CORS policy.
	DisableCORS bool `mapstructure:"disable_cors"`
	// A list of extra origins to allow for CORS.
	CORSOrigins []string `mapstructure:"cors_allowed_origins"`
	// JWT signing key.
	JWTSigningKey string `mapstructure:"jwt_signkey"`
	// JWT expiration in hours.
	JWTExpiration int `mapstructure:"jwt_expiration"`
	// ResetPasswordTokenExp expiration the token expiration
	// for reset password and email confirmation requests in minutes.
	ResetPasswordTokenExp int `mapstructure:"reset_pwd_token_expiration"`
	// Maximum file size to allow for samples.
	MaxFileSize int64 `mapstructure:"max_file_size"`
	// Maximum avatar size to allow for user profile picture.
	MaxAvatarSize int64 `mapstructure:"max_avatar_file_size"`
	// Database configuration.
	DB DatabaseCfg `mapstructure:"db"`
	// Broker server configuration.
	Broker BrokerCfg `mapstructure:"nsq"`
	// Frontend Configuration.
	UI UICfg `mapstructure:"ui"`
	// Object storage configuration.
	ObjStorage StorageCfg `mapstructure:"storage"`
	// SMTP server configuration.
	SMTP SMTPConfig `mapstructure:"smtp"`
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
	env := os.Getenv("SFW_WEBAPIS_DEPLOYMENT_KIND")
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
