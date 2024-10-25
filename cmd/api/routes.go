package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependences) routes() http.Handler {
	//setup a new router
	router := httprouter.New()

	//handle 405
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	//method 404
	router.NotFound = http.HandlerFunc(a.notFoundResponse)

	//setup routes
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthChechHandler)
	router.HandlerFunc(http.MethodPost, "/v1/signin", a.createCommentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/signin/:id", a.displayCommentHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/signin/:id", a.updateCommentHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/signin/:id", a.deleteCommentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/signin", a.listCommentHandler)
	return a.recoverPanic(router)
}
