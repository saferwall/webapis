# saferwall web apis [![GoDoc](http://godoc.org/github.com/saferwall/saferwall-api?status.svg)](https://pkg.go.dev/github.com/saferwall/saferwall-api) [![Report Card](https://goreportcard.com/badge/github.com/saferwall/saferwall-api)](https://goreportcard.com/report/github.com/saferwall/saferwall-api)

## Preface
This repository powers the web service api used in https://saferwall.com.
This web service uses [Echo](https://echo.labstack.com/) as web application framework, [Couchbase](https://www.couchbase.com/) as NoSQL database and [Zap logger](https://pkg.go.dev/go.uber.org/zap) as logger.
This sample application provides only several functions as Web APIs.
Please refer to the 'Service' section about the detail of those functions.

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