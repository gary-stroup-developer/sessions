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
	r.HandleFunc("/logbook", handlers.LogBook)           //view all entries
	r.HandleFunc("/session/entry", handlers.GymSession)  //form to submit workout
	r.HandleFunc("/user/session/", handlers.ViewWorkout) //view entry by id in query param

	r.Handle("/favicon.ico", http.NotFoundHandler())

	fileServer := http.FileServer(http.Dir("./static/"))
	r.Handle("/resources/", http.StripPrefix("/resources/", fileServer))
}
