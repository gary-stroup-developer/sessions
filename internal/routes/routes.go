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
	r.HandleFunc("/entry", handlers.SubmitWorkout)
	r.HandleFunc("/workout/", handlers.ViewWorkout)

	r.Handle("/favicon.ico", http.NotFoundHandler())

	fileServer := http.FileServer(http.Dir("./static/"))
	r.Handle("/resources/", http.StripPrefix("/resources/", fileServer))
}
