package routes

import (
	"gary-stroup-developer/sessions/internal/handlers"
	"net/http"
)

func Routes(r *http.ServeMux) {
	r.HandleFunc("/", handlers.Index)
	r.HandleFunc("/signup", handlers.Signup)
	r.HandleFunc("/signin", handlers.Login)

	r.HandleFunc("/dashboard/", handlers.Dashboard)
	r.HandleFunc("/logbook", handlers.LogBook)          //view all workout entries
	r.HandleFunc("/session/entry", handlers.GymSession) //form to create workout entry and submit to database

	r.HandleFunc("/user/session/edit/", handlers.EditWorkoutEntry)
	r.HandleFunc("/user/session/delete/", handlers.DeleteWorkoutEntry)

	r.Handle("/favicon.ico", http.NotFoundHandler())

	fileServer := http.FileServer(http.Dir("./static/"))
	r.Handle("/resources/", http.StripPrefix("/resources/", fileServer))
}
