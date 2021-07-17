# saferwall web apis [![GoDoc](http://godoc.org/github.com/saferwall/saferwall-api?status.svg)](https://pkg.go.dev/github.com/saferwall/saferwall-api) [![Report Card](https://goreportcard.com/badge/github.com/saferwall/saferwall-api)](https://goreportcard.com/report/github.com/saferwall/saferwall-api) ![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/saferwall/saferwall-api/Build%20&%20Test/main?style=flat-square) [![codecov](https://codecov.io/gh/saferwall/saferwall-api/branch/main/graph/badge.svg?token=KM4B60IL4L)](https://codecov.io/gh/saferwall/saferwall-api)

## Preface
This repository powers the web service api used in https://saferwall.com.

## Vendoring

These packages are used in the project:

- Routing: [Echo](https://echo.labstack.com/)
- Configuration: [viper](github.com/spf13/viper)
- Logging: [zap](https://github.com/uber-go/zap)
- Message queuing: [nsq](github.com/nsqio/go-nsq)
- JSON Web Tokens [jwt](github.com/golang-jwt/jwt)
- Couchbase Driver: [gocb](https://github.com/couchbase/gocb)
- Password Hashing: [bcrypt](https://golang.org/x/crypto/bcrypt)
- Data Validation: [validator](github.com/go-playground/validator)

## Available Endpoints

The following endpoints are available:

- [Get activities](docs/activities/get.md) : `GET /activities/`

## Folder Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

- **build** - contains packaging and Continuous Integration files.
- **cmd** - contains the main function.
- **configs** - contains configuration file templates or default configs.
- **docs** - contains design and user documents.
- **internal** - contains project specific packages with dependencies.
- **pkg** - contains generic packages without project specific dependencies - these can be safely moved to other projects without internal dependencies.

## Improvements compared to the previous implementation

- clean architecture with solid principles.
- Full test coverage
- swagger doc
- Error handling with proper error response generation
- details
    - `File{}` not depending on `peparser`.


## References

- [Standard Package Layout](https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1)
- [How Do You Structure Your Go Apps? by Kat Zień](https://github.com/katzien/go-structure-examples)
- https://www.calhoun.io/moving-towards-domain-driven-design-in-go/
- [Domain Driven Design in Golang - Strategic Design](https://www.damianopetrungaro.com/posts/ddd-using-golang-strategic-design/)
- [Idiometic Go Web Application Structure](http://josebalius.com/posts/go-app-structure/)
- [A clean architecture for Web Application in Go lang](https://medium.com/wesionary-team/a-clean-architecture-for-web-application-in-go-lang-4b802dd130bb)

- https://github.com/golang/go/wiki/CodeReviewComments
- https://golang.org/doc/effective_go
- https://peter.bourgon.org/go-best-practices-2016/
- https://www.youtube.com/watch?v=dp1cc6-QKY0