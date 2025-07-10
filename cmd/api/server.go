package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type user struct {
	Name string `json:"name"`
	Age string `json:"age"`
	City string `json:"city"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello Root Route")
	fmt.Println("Hello Root Route")
}

func teachersHandler(w http.ResponseWriter, r *http.Request) {
	// teachers/{id}
	// teacher/?key=value&query=value2&sortyby=email&sortOrder=ASC
	switch r.Method {
	case http.MethodGet:
		fmt.Println(r.URL.Path)
		path := strings.TrimPrefix(r.URL.Path, "/teachers/")
		userID := strings.TrimSuffix(path, "/")

		fmt.Println("The ID is:", userID)

		fmt.Println("Query Params", r.URL.Query())
		queyParams := r.URL.Query()
		sortby := queyParams.Get("sortby")
		key := queyParams.Get("key")
		sortOrder := queyParams.Get("sortorder")

		if sortOrder == "" {
			sortOrder = "DESC"
		}

		fmt.Printf("Sortby: %v, Sort order: %v, Key: %v", sortby, sortOrder, key)

		fmt.Fprintf(w, "Hello GET Method on Teachers Route")
		// fmt.Println("Hello GET Method on Teachers Route")
		return
	case http.MethodPost:
		fmt.Fprintf(w, "Hello POST Method on Teachers Route")
		fmt.Println("Hello POST Method on Teachers Route")
		return
	case http.MethodPut:
		fmt.Fprintf(w, "Hello PUT Method on Teachers Route")
		fmt.Println("Hello PUT Method on Teachers Route")
		return
	case http.MethodPatch:
		fmt.Fprintf(w, "Hello PATCH Method on Teachers Route")
		fmt.Println("Hello PATCH Method on Teachers Route")
		return
	case http.MethodDelete:
		fmt.Fprintf(w, "Hello DELETE Method on Teachers Route")
		fmt.Println("Hello DELETE Method on Teachers Route")
		return
	}
	// fmt.Fprintf(w, "Hello Teachers Route")
	// fmt.Println("Hello Teachers Route")
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Hello GET Method on Students Route")
		fmt.Println("Hello GET Method on Students Route")
		return
	case http.MethodPost:
		fmt.Fprintf(w, "Hello POST Method on Students Route")
		fmt.Println("Hello POST Method on Students Route")
		return
	case http.MethodPut:
		fmt.Fprintf(w, "Hello PUT Method on Students Route")
		fmt.Println("Hello PUT Method on Students Route")
		return
	case http.MethodPatch:
		fmt.Fprintf(w, "Hello PATCH Method on Students Route")
		fmt.Println("Hello PATCH Method on Students Route")
		return
	case http.MethodDelete:
		fmt.Fprintf(w, "Hello DELETE Method on Students Route")
		fmt.Println("Hello DELETE Method on Students Route")
		return
	}
	fmt.Fprintf(w, "Hello students Route")
	fmt.Println("Hello students Route")
}

func execsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Hello GET Method on Execs Route")
		fmt.Println("Hello GET Method on Execs Route")
		return
	case http.MethodPost:
		fmt.Fprintf(w, "Hello POST Method on Execs Route")
		fmt.Println("Hello POST Method on Execs Route")
		return
	case http.MethodPut:
		fmt.Fprintf(w, "Hello PUT Method on Execs Route")
		fmt.Println("Hello PUT Method on Execs Route")
		return
	case http.MethodPatch:
		fmt.Fprintf(w, "Hello PATCH Method on Execs Route")
		fmt.Println("Hello PATCH Method on Execs Route")
		return
	case http.MethodDelete:
		fmt.Fprintf(w, "Hello DELETE Method on Execs Route")
		fmt.Println("Hello DELETE Method on Execs Route")
		return
	}
	fmt.Fprintf(w, "Hello execs Route")
	fmt.Println("Hello execs Route")
}

func main() {

	port := ":3000"

	http.HandleFunc("/", rootHandler)

	http.HandleFunc("/teachers/", teachersHandler )

	http.HandleFunc("/students/", studentsHandler)

	http.HandleFunc("/execs/", execsHandler)

	fmt.Println("Server is running on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
	
}