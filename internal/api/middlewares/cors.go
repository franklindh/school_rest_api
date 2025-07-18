package middlewares

import (
	"fmt"
	"net/http"
	"slices"
)

// api is hostes at www.myapi.com
// frontend server is at www.myfrontend.com

// Allowed origins
var allowedOrigins = []string {
	// "https://www.myfrontend.com"
	"https://my-origin-url.com",
	"https://localhost:3000",
}

// func Cors(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		origin := r.Header.Get("Origin")
// 		fmt.Println(origin)

// 		if isOriginAllowed(origin) {

// 		} else {
// 			http.Error(w, "Not alllowed by CORS", http.StatusForbidden)
// 			return 
// 		}

// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authotization")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
// 		w.Header().Set("Access-Control-Allow-Credentials", "true")
// 		next.ServeHTTP(w, r)
// 	})
// }

func Cors(next http.Handler) http.Handler {
	fmt.Println("Cors Middleware...")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Cors Middleware being returned...")
			origin := r.Header.Get("Origin")

			if isOriginAllowed(origin) {
					// --- INI BAGIAN YANG BENAR ---
					// 1. Kasih tau browser kalau origin ini diizinkan. Ini yang paling penting!
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// 2. Set header-header lainnya.
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Typo 'Authotization' diperbaiki
					w.Header().Set("Access-Control-Allow-Expose-Headers", "Authorization")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Max-Age", "3600")

					if r.Method == http.MethodOptions {
						return 
					}

					// 3. Lanjutkan ke handler utama.
					next.ServeHTTP(w, r)
					fmt.Println("Cors Middleware ends...")
			} else {
					// Jika tidak diizinkan, tolak.
					http.Error(w, "Not allowed by CORS", http.StatusForbidden)
					return
			}
	})
}

func isOriginAllowed(origin string) bool {
	return slices.Contains(allowedOrigins, origin) 
}