# saferwall Web APIs [![GoDoc](http://godoc.org/github.com/saferwall/saferwall-api?status.svg)](https://pkg.go.dev/github.com/saferwall/saferwall-api) [![Report Card](https://goreportcard.com/badge/github.com/saferwall/saferwall-api)](https://goreportcard.com/report/github.com/saferwall/saferwall-api) ![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/saferwall/saferwall-api/ci.yaml?style=flat-square) [![codecov](https://codecov.io/gh/saferwall/saferwall-api/branch/main/graph/badge.svg?token=KM4B60IL4L)](https://codecov.io/gh/saferwall/saferwall-api)

## Preface
This repository powers the web service API used in https://saferwall.com.

## Vendoring

These packages are used in the project:

- Language: [Golang 1.21+](https://go.dev/)
- Routing: [Echo](https://echo.labstack.com/)
- Configuration: [viper](github.com/spf13/viper)
- Logging: [zap](https://github.com/uber-go/zap)
- Message queuing: [nsq](github.com/nsqio/go-nsq)
- Authentication [jwt](github.com/golang-jwt/jwt)
- Database: [gocb](https://github.com/couchbase/gocb)
- Password Hashing: [bcrypt](https://golang.org/x/crypto/bcrypt)
- Data Validation: [validator](github.com/go-playground/validator)
- i18n Translator: [universal-translator](github.com/go-playground/universal-translator)

## Folder Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

- **build** - contains packaging and Continuous Integration files.
- **cmd** - contains the main function.
- **configs** - contains configuration file templates or default configs.
- **docs** - contains design and user documents.
- **db** - contains database sql-like (n1ql) queries.
- **internal** - contains private project specific code.
- **pkg** - contains generic packages without project specific dependencies - these can be safely moved to other projects without internal dependencies.

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
