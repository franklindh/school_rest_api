package handlers

import (
	"fmt"
	"net/http"
)

func Staff(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Hello GET Method on Staff Route")
		return
	case http.MethodPost:
		fmt.Fprintf(w, "Hello POST Method on Staff Route")
		return
	case http.MethodPut:
		fmt.Fprintf(w, "Hello PUT Method on Staff Route")
		return
	case http.MethodPatch:
		fmt.Fprintf(w, "Hello PATCH Method on Staff Route")
		return
	case http.MethodDelete:
		fmt.Fprintf(w, "Hello DELETE Method on Staff Route")
		return
	}
}
