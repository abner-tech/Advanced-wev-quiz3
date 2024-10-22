package main

import (
	"fmt"
	"net/http"
)

func (a *applicationDependences) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//defer will be called when the stack unwinds
		defer func() {
			//recover from panic
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "Close")
				a.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
