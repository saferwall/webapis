################################
# STEP 1 build executable binary
################################

FROM golang:1.17-alpine AS build-stage

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata \
    && update-ca-certificates 2>/dev/null || true

# Set the Current Working Directory inside the container
WORKDIR /webapi

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies.
# Dependencies will be cached if the go.mod and go.sum files are not changed.
RUN go mod download

# Copy the source from the current directory to the Working Directory inside
# the container
COPY . .

# Build the go app.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static"' -o /go/bin/server cmd/main.go

############################
# STEP 2 build a small image
############################

FROM alpine:latest
LABEL maintainer="https://github.com/saferwall/saferwall-api"
LABEL version="0.0.3"
LABEL description="Saferwall web APIs service"

ENV USER saferwall
ENV GROUP saferwall

# Set the Current Working Directory inside the container.
WORKDIR /saferwall

# Copy our static executable.
COPY --from=build-stage /go/bin/server .

# Copy the config files.
COPY configs/ conf/
COPY db/ db/
COPY templates/ templates/

# Create an app user so our program doesn't run as root.
RUN addgroup -g 102 -S $GROUP \
    && adduser -u 101 -S $USER -G $GROUP \
    && chown -R $USER:$GROUP /saferwall

# Switch to our user.
USER saferwall

ENTRYPOINT ["/saferwall/server", "-config", "/saferwall/conf",\
 "-db", "/saferwall/db", "-tpl", "/saferwall/templates"]
