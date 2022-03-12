// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/saferwall/saferwall-api/internal/entity"
	e "github.com/saferwall/saferwall-api/internal/errors"
)

const (
	jwtCookieName = "JWTCookie"
)

// Handler returns a JWT-based authentication middleware.
func Handler(verificationKey string) echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:     []byte(verificationKey),
		SuccessHandler: successHandler,
		ParseTokenFunc: parseTokenFunc([]byte(verificationKey)),
		ErrorHandler:   errorHandler,
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
		token.Claims.(jwt.MapClaims)["isAdmin"].(bool),
	)
	c.SetRequest(c.Request().WithContext(ctx))

	// determines the source of the API request
	src := reqSource(c.Request().UserAgent())
	ctx = WithSource(c.Request().Context(), src)
	c.SetRequest(c.Request().WithContext(ctx))
}

func errorHandler(err error) error {
	return e.Unauthorized("invalid or expired jwt")
}

// IsAuthenticated middleware checks if a user is authenticated.
// If not, it calls the next handler HTTP.
// If yes, it validates the JWT token and returns an error if token is invalid.
func IsAuthenticated(authHandler echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// check first if token was handed by a cookie.
			authScheme := "Bearer"
			_, err := c.Cookie(jwtCookieName)
			if err == nil {
				return authHandler(next)(c)
			}

			// if not, check in Authorization header.
			auth := c.Request().Header.Get("Authorization")
			l := len(authScheme)
			if len(auth) > l+1 && auth[:l] == authScheme {
				return authHandler(next)(c)
			}

			return next(c)
		}
	}
}

// WithUser returns a context that contains the user identity from the given JWT.
func WithUser(ctx context.Context, id string, isAdmin bool) context.Context {
	return context.WithValue(
		ctx, entity.UserKey, entity.User{
			Username: id,
			Admin:    isAdmin})
}

// CurrentUser returns the user identity from the given context.
// Nil is returned if no user identity is found in the context.
func CurrentUser(ctx context.Context) Identity {
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		return user
	}
	return nil
}

// WithSource returns a context that contains the source of the HTTP request.
func WithSource(ctx context.Context, src string) context.Context {
	return context.WithValue(ctx, entity.SourceKey, src)
}

// isBrowser returns true when the HTTP request is coming from a known user agent.
func isBrowser(userAgent string) bool {
	browserList := []string{
		"Chrome", "Chromium", "Mozilla", "Opera", "Safari", "Edge", "MSIE",
	}

	for _, browserName := range browserList {
		if strings.Contains(userAgent, browserName) {
			return true
		}
	}
	return false
}

func reqSource(userAgent string) string {
	if isBrowser(userAgent) {
		return "web"
	} else {
		return "api"
	}
}
