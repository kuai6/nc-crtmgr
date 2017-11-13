package main

import "github.com/julienschmidt/httprouter"

func InitRouter() *httprouter.Router {
	router := httprouter.New()
	router.POST("/api/v1/generate", GenerateHandler)
	router.POST("/api/v1/validate", ValidateHandler)

	return router
}
