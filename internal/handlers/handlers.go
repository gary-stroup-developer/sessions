package handlers

import (
	"database/sql"
	"gary-stroup-developer/sessions/internal/models"
	"gary-stroup-developer/sessions/internal/sessions"
	"html/template"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Repository struct {
	Template *template.Template
	DB       *sql.DB
	DbUsers  map[string]models.UserInfo // session ID, user
}

func NewRepo(db *sql.DB, tpl *template.Template) *Repository {
	return &Repository{
		Template: tpl,
		DB:       db,
		DbUsers:  make(map[string]models.UserInfo),
	}
}

var Repo *Repository

func SetRepo(r *Repository) {
	Repo = r
}

func Index(w http.ResponseWriter, req *http.Request) {

	Repo.Template.ExecuteTemplate(w, "index.gohtml", nil)
}

func Dashboard(w http.ResponseWriter, req *http.Request) {

	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	u := sessions.GetUser(req, Repo.DbUsers)

	Repo.Template.ExecuteTemplate(w, "dashboard.gohtml", u)
}

func Signup(w http.ResponseWriter, req *http.Request) {
	if sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {
		//get values submitted by form
		un := req.FormValue("username")
		p := req.FormValue("password")
		f := req.FormValue("firstname")
		l := req.FormValue("lastname")

		//hash password submitted & check for errors
		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		//send data to be inserted into database & check for error
		u, err := signUserUp(un, bs, f, l)

		if err != nil {
			http.Error(w, "Uh oh. Something went wrong on our end. Try signing up again", http.StatusInternalServerError)
			return
		}

		//create a session cookie
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		//store cookie in browser
		http.SetCookie(w, c)

		//bind session cookie with user
		Repo.DbUsers[c.Value] = u

		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "signup.gohtml", nil)
}

func Login(w http.ResponseWriter, req *http.Request) {
	if sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {
		//get the form values
		un := req.FormValue("username")
		p := req.FormValue("password")

		u, err := logUserIn(w, un, p)

		if err != nil {
			log.Fatalln(err)
			return
		}

		//create a cookie
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}

		http.SetCookie(w, c)

		Repo.DbUsers[c.Value] = u

		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "login.gohtml", nil)
}

func GymSession(w http.ResponseWriter, req *http.Request) {
	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {
		//parse th form data
		req.ParseForm()

		//parse each field into []Workout
		wkout, err := logWorkout(req.Form["description"], req.Form["sets"], req.Form["reps"])

		if err != nil {
			http.Error(w, "workout not logged in bro!", http.StatusBadRequest)
		}
		log.Println(wkout)

		//need to create function to insert wkout into database
		// //send user info to be stored in database
		// err = signUserUp(data []Workout)

		// if err != nil {
		// 	log.Fatalln(err)
		// 	return
		// }

		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "entry.gohtml", nil)
}

func ViewWorkout(w http.ResponseWriter, req *http.Request) {
	//Step 1: check to see if logged in

	//Step 2: get id from url
	// id, _ := url.Parse("http://localhost:8080/workout/?id=55")
	userID := req.URL.Query()["id"][0]
	log.Println(userID)
	//Step 3: search database for workout with that id

	//Step 4: send data to template
	Repo.Template.ExecuteTemplate(w, "viewEntry.gohtml", nil)
}
