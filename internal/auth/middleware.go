// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Handler returns a JWT-based authentication middleware.
func Handler(verificationKey string) echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(verificationKey),
		TokenLookup: "header:Authorization,cookie:JWTCookie",
	})
}
