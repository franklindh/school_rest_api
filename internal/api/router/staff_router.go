package router

import (
	"net/http"
	"restapi/internal/api/handlers"
)

func StaffRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /staff", handlers.GetStaffHandler)
	mux.HandleFunc("POST /staff", handlers.AddStaffHandler)
	mux.HandleFunc("PATCH /staff", handlers.PatchStaffHandler)

	mux.HandleFunc("GET /staff/{id}", handlers.GetOneStaffHandler)
	mux.HandleFunc("PATCH /staff/{id}", handlers.PatchOneStaffHandler)
	mux.HandleFunc("DELETE /staff/{id}", handlers.DeleteOneStaffHandler)
	mux.HandleFunc("POST /staff/{id}/updatedpassword", handlers.UpdatePasswordHandler)

	mux.HandleFunc("POST /staff/login", handlers.LoginHandler)
	mux.HandleFunc("POST /staff/logout", handlers.LogoutHandler)
	mux.HandleFunc("POST /staff/forgotpassword", handlers.ForgotPasswordHandler)
	mux.HandleFunc("POST /staff/resetpassword/reset/{resetcode}", handlers.GetStaffHandler)

	return mux
}
