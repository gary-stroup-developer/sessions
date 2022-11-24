package routes

import (
	"gary-stroup-developer/sessions/internal/handlers"
	"net/http"
)

func Routes(r *http.ServeMux) {
	r.HandleFunc("/", handlers.Index)
	r.HandleFunc("/dashboard", handlers.Dashboard)
	r.HandleFunc("/signup", handlers.Signup)
	r.HandleFunc("/signin", handlers.Login)

	r.Handle("/favicon.ico", http.NotFoundHandler())
}
