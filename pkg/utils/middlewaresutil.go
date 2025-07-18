package utils

import "net/http"

//Middleware is a function that wraps an http.Handler with additional functinality
type Middelware func(http.Handler) http.Handler

func ApplyMiddlewares(handler http.Handler, middlewares ...Middelware) http.Handler {
for _, middleware := range middlewares {
	handler = middleware(handler)
}
return handler
}