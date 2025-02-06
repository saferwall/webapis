// Copyright 2018 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package file

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/saferwall/saferwall-api/internal/entity"
	"github.com/saferwall/saferwall-api/internal/errors"
	"github.com/saferwall/saferwall-api/pkg/log"
	"github.com/saferwall/saferwall-api/pkg/pagination"
)

const (
	KB = 1000
	MB = 1000 * KB
)

type resource struct {
	service       Service
	logger        log.Logger
	maxSampleSize int64 // expressed in bytes.
}

func RegisterHandlers(g *echo.Group, service Service, logger log.Logger,
	maxFileSize int,
	requireLogin, optionalLogin, verifyHash, cacheResponse, modifyResponse echo.MiddlewareFunc) {

	res := resource{service, logger, int64(maxFileSize * MB)}

	g.GET("/files/", res.list, requireLogin)
	g.POST("/files/", res.create, requireLogin)
	g.HEAD("/files/:sha256/", res.exists, verifyHash)
	g.GET("/files/:sha256/", res.get, modifyResponse, cacheResponse, verifyHash)
	g.PUT("/files/:sha256/", res.update, verifyHash, requireLogin)
	g.PATCH("/files/:sha256/", res.patch, verifyHash, requireLogin)
	g.DELETE("/files/:sha256/", res.delete, verifyHash, requireLogin)
	g.GET("/files/:sha256/strings/", res.strings, modifyResponse, cacheResponse, verifyHash)
	g.GET("/files/:sha256/summary/", res.summary, modifyResponse, cacheResponse, verifyHash)
	g.GET("/files/:sha256/comments/", res.comments, modifyResponse, verifyHash, optionalLogin)
	g.POST("/files/:sha256/like/", res.like, verifyHash, requireLogin)
	g.POST("/files/:sha256/unlike/", res.unlike, verifyHash, requireLogin)
	g.POST("/files/:sha256/rescan/", res.rescan, verifyHash, requireLogin)
	g.GET("/files/:sha256/download/", res.download, verifyHash, requireLogin)
	g.GET("/files/:sha256/generate-presigned-url/", res.generatePresignedURL, verifyHash, requireLogin)
	g.GET("/files/:sha256/meta-ui/", res.metaUI, verifyHash, optionalLogin)
	g.POST("/files/search/", res.search, requireLogin)
	g.GET("/files/search/autocomplete?&goodware", res.search, requireLogin)
	//
}

// @Summary Check if a file exists.
// @Description Check weather a file exists in the database.
// @Tags File
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} entity.File
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256} [head]
func (r resource) exists(c echo.Context) error {

	ctx := c.Request().Context()
	exists, err := r.service.Exists(ctx, c.Param("sha256"))
	if err != nil {
		return err
	}

	if exists {
		return c.NoContent(http.StatusOK)
	}

	return c.NoContent(http.StatusNotFound)
}

// @Summary Get a file report
// @Description Retrieves the content of a file report.
// @Tags File
// @Accept json
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} entity.File
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256} [get]
func (r resource) get(c echo.Context) error {

	// the `fields` query parameter is used to limit the fields
	// to include in the response.
	var fields []string
	if fieldsParam := c.QueryParam("fields"); fieldsParam != "" {
		fields = strings.Split(fieldsParam, ",")
	}

	ctx := c.Request().Context()
	file, err := r.service.Get(ctx, c.Param("sha256"), fields)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, file)
}

// @Summary Submit a new file for scanning
// @Description Upload file for analysis.
// @Tags File
// @Accept mpfd
// @Produce json
// @Param file formData file true  "binary file"
// @Success 201 {object} entity.File
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 413 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/ [post]
// @Security Bearer
func (r resource) create(c echo.Context) error {

	ctx := c.Request().Context()
	f, err := c.FormFile("file")
	if err != nil {
		r.logger.With(ctx).Info(err)
		return errors.BadRequest("missing file in form request")
	}

	if f.Size > r.maxSampleSize {
		r.logger.With(ctx).Info("payload too large")
		return errors.TooLargeEntity("")
	}

	src, err := f.Open()
	if err != nil {
		r.logger.With(ctx).Error(err)
		return errors.InternalServerError("")
	}
	defer src.Close()

	r.logger.With(ctx).Debug("Header is: %v", c.Request().Header)

	var scanCfg FileScanRequest
	if err := c.Bind(&scanCfg); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return errors.BadRequest("")
	}

	input := CreateFileRequest{
		src:      src,
		filename: f.Filename,
		geoip:    c.Request().Header.Get("X-Geoip-Country"),
		scanCfg:  scanCfg,
	}
	file, err := r.service.Create(ctx, input)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, file)
}

// @Summary Update a file report (full update)
// @Description Replace a file report with a new report
// @Tags File
// @Accept json
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} entity.File
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256} [put]
// @Security Bearer
func (r resource) update(c echo.Context) error {

	var isAdmin bool
	ctx := c.Request().Context()
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		isAdmin = user.IsAdmin()
	}
	if !isAdmin {
		return errors.Forbidden("")
	}

	var input UpdateFileRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return errors.BadRequest("")
	}

	file, err := r.service.Update(ctx, c.Param("sha256"), input)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, file)
}

// @Summary Update a file report (partial update)
// @Description Patch a portion of a file report.
// @Tags File
// @Accept json
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} entity.File
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256} [patch]
// @Security Bearer
func (r resource) patch(c echo.Context) error {
	var isAdmin bool
	ctx := c.Request().Context()
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		isAdmin = user.IsAdmin()
	}
	if !isAdmin {
		return errors.Forbidden("")
	}
	return nil
}

// @Summary Deletes a file
// @Description Deletes a file by ID.
// @Tags File
// @Accept json
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 204 {object} entity.File
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256} [delete]
// @Security Bearer
func (r resource) delete(c echo.Context) error {

	var isAdmin bool
	ctx := c.Request().Context()
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		isAdmin = user.IsAdmin()
	}
	if !isAdmin {
		return errors.Forbidden("")
	}

	file, err := r.service.Delete(ctx, c.Param("sha256"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, file)
}

// @Summary Retrieves a paginated list of files
// @Description List files
// @Tags File
// @Accept json
// @Produce json
// @Param per_page query uint false "Number of files per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages{items=[]entity.File}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/ [get]
// @Security Bearer
func (r resource) list(c echo.Context) error {
	var isAdmin bool
	ctx := c.Request().Context()
	if user, ok := ctx.Value(entity.UserKey).(entity.User); ok {
		isAdmin = user.IsAdmin()
	}
	if !isAdmin {
		return errors.Forbidden("")
	}

	// the `fields` query parameter is used to limit the fields
	// to include in the response.
	var fields []string
	if fieldsParam := c.QueryParam("fields"); fieldsParam != "" {
		fields = strings.Split(fieldsParam, ",")
	}

	if len(fields) > 0 {
		allowed := areFieldsAllowed(fields)
		if !allowed {
			return errors.BadRequest("field not allowed")
		}
	}

	count, err := r.service.Count(ctx)
	if err != nil {
		return err
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	files, err := r.service.Query(ctx, pages.Offset(), pages.Limit(), fields)
	if err != nil {
		return err
	}
	pages.Items = files
	return c.JSON(http.StatusOK, pages)
}

// @Summary Searches files based on files' metadata
// @Description Search files
// @Tags File
// @Accept json
// @Produce json
// @Param per_page query uint false "Number of files per page"
// @Param page query uint false "Specify the page number"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/search/ [post]
// @Security Bearer
func (r resource) search(c echo.Context) error {
	ctx := c.Request().Context()

	var input FileSearchRequest
	if err := c.Bind(&input); err != nil {
		r.logger.With(c.Request().Context()).Info(err)
		return errors.BadRequest("")
	}

	search, err := r.service.Search(ctx, input)
	if err != nil {
		return err
	}

	pages := pagination.New(input.Page, input.PerPage, int(search.TotalHits))
	pages.Items = search.Results
	return c.JSON(http.StatusOK, pages)
}

// @Summary Returns a paginated list of strings
// @Description List strings of a file.
// @Tags File
// @Accept json
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Param per_page query uint false "Number of strings per page"
// @Param page query uint false "Specify the page number"
// @Param q query string false "Specify the string to search for"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256}/strings/ [get]
func (r resource) strings(c echo.Context) error {
	ctx := c.Request().Context()

	searchTerm := c.QueryParam("q")

	count, err := r.service.CountStrings(ctx, c.Param("sha256"), searchTerm)
	if err != nil {
		return err
	}

	// The offset and limit passed to the repository interface are accessing
	// an array [limit:off] rather than OFFSET/LIMIT in SQL.
	// Adjust them so they can access arrays properly.
	pages := pagination.NewFromRequest(c.Request(), count)
	limit := pages.Offset() + pages.Limit()
	if count < limit {
		limit = count
	}

	strings, err := r.service.Strings(
		ctx, c.Param("sha256"), searchTerm, pages.Offset(), limit)
	if err != nil {
		return err
	}
	pages.Items = strings
	return c.JSON(http.StatusOK, pages)
}

// @Summary File summary and metadata
// @Description File metadata returned in the summary view of a file.
// @Tags File
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256}/summary/ [get]
// @Security Bearer || {}
func (r resource) summary(c echo.Context) error {
	ctx := c.Request().Context()
	fileSummary, err := r.service.Summary(ctx, c.Param("sha256"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, fileSummary)
}

// @Summary Returns a paginated list of file comments
// @Description List of comments for a given file.
// @Tags File
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} pagination.Pages
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256}/comments/ [get]
// @Security Bearer || {}
func (r resource) comments(c echo.Context) error {
	ctx := c.Request().Context()

	count, err := r.service.CountComments(ctx, c.Param("sha256"))
	if err != nil {
		return err
	}
	pages := pagination.NewFromRequest(c.Request(), count)
	comments, err := r.service.Comments(
		ctx, c.Param("sha256"), pages.Offset(), pages.Limit())
	if err != nil {
		return err
	}
	pages.Items = comments
	return c.JSON(http.StatusOK, pages)
}

// @Summary Like a file
// @Description Adds a file to the like list.
// @Tags File
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} object{}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256}/like/ [post]
// @Security Bearer
func (r resource) like(c echo.Context) error {
	ctx := c.Request().Context()
	err := r.service.Like(ctx, c.Param("sha256"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}

// @Summary Unlike a file
// @Description Removes a file from the like list.
// @Tags File
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} object{}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256}/unlike/ [post]
// @Security Bearer
func (r resource) unlike(c echo.Context) error {
	ctx := c.Request().Context()
	err := r.service.Unlike(ctx, c.Param("sha256"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}

// @Summary Rescan an existing file
// @Description Rescan an existing file.
// @Tags File
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} object{}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256}/rescan/ [post]
// @Security Bearer
func (r resource) rescan(c echo.Context) error {
	ctx := c.Request().Context()

	var scanCfg FileScanRequest
	if c.Request().ContentLength > 0 {
		if err := c.Bind(&scanCfg); err != nil {
			r.logger.With(c.Request().Context()).Info(err)
			return errors.BadRequest("")
		}
	}

	err := r.service.ReScan(ctx, c.Param("sha256"), scanCfg)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}{"ok", http.StatusOK})
}

// @Summary Download a file
// @Description Download a binary file. Files are in zip format and password protected.
// @Tags File
// @Produce mpfd
// @Param sha256 path string true "File SHA256"
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256}/download/ [get]
// @Security Bearer
func (r resource) download(c echo.Context) error {
	ctx := c.Request().Context()
	var zippedFile string
	err := r.service.Download(ctx, c.Param("sha256"), &zippedFile)
	if err != nil {
		switch err {
		case ErrObjectNotFound:
			return errors.NotFound("")
		default:
			return err
		}
	}
	return c.File(zippedFile)
}

// @Summary Generate a pre-signed URL for downloading samples.
// @Description Generate a pre-signed URL to download samples directly from the object storage.
// @Tags File
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} object{}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256}/generate-presigned-url/ [post]
// @Security Bearer
func (r resource) generatePresignedURL(c echo.Context) error {
	ctx := c.Request().Context()
	preSignedURL, err := r.service.GeneratePresignedURL(ctx, c.Param("sha256"))
	if err != nil {
		switch err {
		case ErrObjectNotFound:
			return errors.NotFound("")
		default:
			return err
		}
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
		URL     string `json:"url"`
	}{"ok", http.StatusOK, preSignedURL})
}

// @Summary Frontend Metadata
// @Description Frontend metadata fields such as navbar menus.
// @Tags File
// @Produce json
// @Param sha256 path string true "File SHA256"
// @Success 200 {object} object{}
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /files/{sha256}/meta-ui/ [get]
// @Security Bearer || {}
func (r resource) metaUI(c echo.Context) error {
	ctx := c.Request().Context()
	fileSummary, err := r.service.MetaUI(ctx, c.Param("sha256"))
	if err != nil {
		return err
	}

	// Cache when file scanning is finished.
	summary, ok := fileSummary.(map[string]interface{})
	if ok {
		status, statusOK := summary["status"].(entity.FileScanProgressType)
		if statusOK {
			if status == entity.FileScanProgressFinished {
				c.Response().Header().Set("Cache-Control", "public, max-age=300")

			}
		}
	}

	return c.JSON(http.StatusOK, fileSummary)
}
