// Package httputil holds tiny helpers so every handler returns responses in the
// SAME shape (Approach A: raw data on success, a structured error on failure).
//
//	Success:  return the resource directly        -> OK(c, profile)
//	Created:  return the new resource with 201     -> Created(c, post)
//	List:     wrap with pagination metadata        -> List(c, items, meta)
//	No body:  return 204                            -> NoContent(c)
//	Error:    { "error": { "code", "message" } }    -> Error(c, 400, "code", "msg")
//
// The error `code` is a STABLE machine-readable string (e.g. "invalid_code")
// the mobile app can branch on / translate. The `message` is for humans/logs.
package httputil

import "github.com/gin-gonic/gin"

// Meta is pagination info attached to list responses.
type Meta struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// errorBody is the consistent error envelope.
type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK sends 200 with the resource as-is.
func OK(c *gin.Context, data any) {
	c.JSON(200, data)
}

// Created sends 201 with the newly created resource.
func Created(c *gin.Context, data any) {
	c.JSON(201, data)
}

// NoContent sends 204 (used for deletes / actions with no body).
func NoContent(c *gin.Context) {
	c.Status(204)
}

// List sends 200 with { "data": [...], "meta": {...} } for paginated lists.
func List(c *gin.Context, data any, meta Meta) {
	c.JSON(200, gin.H{"data": data, "meta": meta})
}

// Error sends the given HTTP status with a { "error": { code, message } } body.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": errorBody{Code: code, Message: message}})
}

// Abort is like Error but also stops the middleware chain (use in middleware).
func Abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": errorBody{Code: code, Message: message}})
}
