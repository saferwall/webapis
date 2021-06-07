// Copyright 2021 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package auth

// import (
// 	"github.com/labstack/echo/v4"
// 	"github.com/qiangxue/go-rest-api/internal/errors"
// 	"github.com/saferwall/saferwall-api/pkg/log"
// )

// // RegisterHandlers registers handlers for different HTTP requests.
// func RegisterHandlers(rg *routing.RouteGroup, service Service, logger log.Logger) {
// 	rg.Post("/login", login(service, logger))
// }

// // login returns a handler that handles user login request.
// func login(service Service, logger log.Logger) routing.Handler {
// 	return func(c *routing.Context) error {
// 		var req struct {
// 			Username string `json:"username"`
// 			Password string `json:"password"`
// 		}

// 		if err := c.Read(&req); err != nil {
// 			logger.With(c.Request.Context()).Errorf("invalid request: %v", err)
// 			return errors.BadRequest("")
// 		}

// 		token, err := service.Login(c.Request.Context(), req.Username, req.Password)
// 		if err != nil {
// 			return err
// 		}
// 		return c.Write(struct {
// 			Token string `json:"token"`
// 		}{token})
// 	}
// }
