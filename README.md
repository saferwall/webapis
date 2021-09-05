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

- `GET /healthcheck/` [health check](docs/healthcheck/get.md]).

## Activities

- `GET /v1/activities/`: [returns a paginated list of activities](docs/activities/get.md).
- `DELETE /v1/activities/`: [deletes multiple activities](docs/activities/delete.md).

### Authentication

- `POST /v1/auth/login/`: [authenticates a user and generates a JWT](docs/auth/login.md).
- `POST /v1/auth/verify/`: [confirms an account from email token](docs/auth/confirm.md).
- `POST /v1/auth/resend-confirmation/`: [re-send confirmation email](docs/auth/resend-confirmation.md).
- `POST /v1/auth/reset-password/`: [send an email with a reset password token](docs/auth/reset-password.md).

### Users resource

- `GET /v1/users/`: [get a paginated list of users](docs/users/get.md).
- `POST /v1/users/`: [creates a new user](docs/users/post.md).
- `PATCH /v1/users/`: [update multiple users](docs/users/patch.md).
- `PUT /v1/users/`: [replace multiple users](docs/users/put.md).
- `DELETE /v1/users/`: [deletes multiple users](docs/users/delete.md).

### User resource

- `GET /v1/users/:username/`: [get a detailed information of a user](docs/user/get.md).
- `PATCH /v1/users/:username/`: [updates an existing user](docs/user/patch.md).
- `PUT /v1/users/:username/`: [replaces an existing user](docs/user/post.md).
- `DELETE /v1/users/:username/`: [deletes an existing user](docs/user/delete.md).
- `PATCH /v1/users/:username/password/`: [updates an authenticated user's password](docs/auth/patch.md).
- `GET /v1/users/:username/likes/`: [get a paginated list of user's likes](docs/user/get.md).
- `GET /v1/users/:username/submissions/`: [get a paginated list of user's submissions](docs/profile/submissions.md).
- `GET /v1/users/:username/following/`: [get a paginated list of user's following](docs/profile/following.md).
- `GET /v1/users/:username/followers/`: [get a paginated list of user's followers](docs/profile/followers.md).
- `GET /v1/users/:username/comments/`: [get a paginated list of user's comments](docs/profile/comments.md).
- `POST /v1/users/:username/avatar/`: [updates an existing user's avatar](docs/user/avatar.md).

### User actions

- `POST /v1/users/:username/follow/`: [follow an existing user](docs/actions/follow.md).
- `POST /v1/users/:username/unfollow/`: [unfollow an existing user](docs/actions/unfollow.md).

### Files resource

- `GET /v1/files/`: [get a paginated list of files](docs/files/get.md).
- `POST /v1/files/`: [creates a new file](docs/files/post.md).
- `PATCH /v1/files/`: [update multiple files](docs/files/patch.md).
- `PUT /v1/files/`: [replace multiple files](docs/files/put.md).
- `DELETE /v1/files/`: [deletes multiple files](docs/files/delete.md).

### File resource

- `GET /v1/files/:sha256/`: [get a detailed information of a file](docs/file/get.md).
- `PATCH /v1/files/:sha256/`: [updates an existing file](docs/file/patch.md).
- `PUT /v1/files/:sha256/`: [replaces an existing file](docs/file/post.md).
- `DELETE /v1/files/:sha256/`: [deletes an existing file](docs/file/delete.md).

### File actions

- `POST /v1/files/:sha256/like/`: [likes an existing file](docs/actions/like.md).
- `POST /v1/files/:sha256/unlike/`: [unlike an existing file](docs/actions/unlike.md).
- `POST /v1/files/:sha256/rescan/`: [re-scan an existing file](docs/actions/rescan.md).
- `GET /v1/files/:sha256/download/`: [downloads an existing file](docs/actions/download.md).

### Comments resource

- `GET /v1/comments/` [returns a paginated list of comments](docs/comments/get.md).
- `POST /v1/comments/`: [creates a new comment](docs/users/post.md).
- `DELETE /v1/comments/` [deletes multiple comments](docs/comments/delete.md).

### Comment resource

- `GET /v1/comments/:id/`: [get a detailed information of a comment](docs/comment/get.md).
- `PATCH /v1/comments/:id/`: [updates an existing comment](docs/comment/patch.md).
- `PUT /v1/comments/:id/`: [replaces an existing comment](docs/comment/post.md).
- `DELETE /v1/comments/:id/`: [deletes an existing comment](docs/comment/delete.md).

Rules for mapping HTTP methods to CRUD:

```http
POST   - Create (add record into database)
GET    - Read (get record from the database)
PATCH  - Update (edit record in the database)
PUT    - Replace (replace record in the database)
DELETE - Delete (remove record from the database)
```

Rules for HTTP status codes:

```http
* Create something            - 201 (Created)
* Read something              - 200 (OK)
* Update something            - 200 (OK)
* Delete something            - 200 (OK)
* Missing request information - 400 (Bad Request)
* Unauthorized operation      - 401 (Unauthorized)
* Any other error             - 500 (Internal Server Error)
```

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
- details:
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