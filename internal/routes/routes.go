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
	r.HandleFunc("/logbook", handlers.LogBook)            //view all workout entries
	r.HandleFunc("/session/entry", handlers.GymSession)   //form to create workout entry and submit to database
	r.HandleFunc("/user/session/", handlers.WorkoutEntry) //Read, edit, or delete workout entry

	r.Handle("/favicon.ico", http.NotFoundHandler())

	fileServer := http.FileServer(http.Dir("./static/"))
	r.Handle("/resources/", http.StripPrefix("/resources/", fileServer))
}
