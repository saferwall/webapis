// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saferwall/saferwall-api/internal/entity"
)

// Handler returns a JWT-based authentication middleware.
func Handler(verificationKey string) echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:     []byte(verificationKey),
		SuccessHandler: successHandler,
		ParseTokenFunc: parseTokenFunc([]byte(verificationKey)),
		TokenLookup:    "header:Authorization,cookie:JWTCookie",
	})
}

func parseTokenFunc(signingKey []byte) func(auth string, c echo.Context) (interface{}, error) {
	return func(auth string, c echo.Context) (interface{}, error) {

		keyFunc := func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != "HS256" {
				return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
			}
			return signingKey, nil
		}

		// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
		token, err := jwt.Parse(auth, keyFunc)
		if err != nil {
			return nil, err
		}
		if !token.Valid {
			return nil, errors.New("invalid token")
		}
		return token, nil
	}
}

func successHandler(c echo.Context) {
	token := c.Get("user").(*jwt.Token)
	ctx := WithUser(
		c.Request().Context(),
		token.Claims.(jwt.MapClaims)["id"].(string),
		token.Claims.(jwt.MapClaims)["name"].(string),
	)
	c.SetRequest( c.Request().WithContext(ctx))
}

type contextKey int

const (
	userKey contextKey = iota
)

// WithUser returns a context that contains the user identity from the given JWT.
func WithUser(ctx context.Context, id, name string) context.Context {
	return context.WithValue(ctx, userKey, entity.User{Username: id, FullName: name})
}

// CurrentUser returns the user identity from the given context.
// Nil is returned if no user identity is found in the context.
func CurrentUser(ctx context.Context) Identity {
	if user, ok := ctx.Value(userKey).(entity.User); ok {
		return user
	}
	return nil
}

// JWTCookie creates a cookie to store the JWT token.
func JWTCookie(token string, domain string, expiration int) http.Cookie {
	cookie := http.Cookie{}
	cookie.Name = "JWTCookie"
	cookie.Value = token
	cookie.Expires = time.Now().Add(time.Hour * time.Duration(expiration))
	cookie.Path = "/"
	cookie.Domain = domain
	// cookie.HttpOnly = false
	// cookie.SameSite = http.SameSiteLaxMode
	// cookie.Secure = false
	return cookie
}
