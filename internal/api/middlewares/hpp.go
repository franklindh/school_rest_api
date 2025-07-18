package middlewares

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type HPPOptions struct {
	CheckQuery 									bool
	CheckBody 									bool
	CheckBodyOnlyForContentType string
	Whitelist 									[]string
}

func Hpp(options HPPOptions) func (http.Handler) http.Handler {
	fmt.Println("HPP Middleware...")
	return func (next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("HPP Middleware being returned...")
			if options.CheckBody && r.Method == http.MethodPost && isCorrectContentType(r, options.CheckBodyOnlyForContentType){
				// Filter the body params
				filterBodyParams(r, options.Whitelist)
			}
			if options.CheckQuery && r.URL.Query() != nil {
				// Filter the query params
				filterQueryParams(r, options.Whitelist)
			}
			next.ServeHTTP(w, r)
			fmt.Println("HPP Middleware ends...")
		})
	}
}

func isCorrectContentType(r *http.Request, contentType string) bool {
	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func filterBodyParams(r *http.Request, whitelist []string) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}

	for key, value := range r.Form {
		if len(value) > 1 {
			r.Form.Set(key, value[0]) // first value
			// r.Form.Set(key, value[len(value) - 1]) // last value
		}
		if !isWhiteListed(key, whitelist) {
			delete(r.Form, key)
		}
	}
}

func filterQueryParams(r *http.Request, whitelist []string) {
	query := r.URL.Query()

	for key, value := range query{
		if len(value) > 1 {
			// query.Set(key, value[0]) // first value
			query.Set(key, value[len(value) - 1]) // last value
		}
		if !isWhiteListed(key, whitelist) {
			query.Del(key)
		}
	}
	r.URL.RawQuery = query.Encode()
}
func isWhiteListed(param string, whitelist []string) bool {
	return slices.Contains(whitelist, param)
	// for _, value := range whitelist {
	// 	if param == value {
	// 		return true
	// 	}
	// }
	// return false
}