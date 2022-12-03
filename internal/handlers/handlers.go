package handlers

import (
	"database/sql"
	"gary-stroup-developer/sessions/internal/models"
	"gary-stroup-developer/sessions/internal/sessions"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	uuid "github.com/satori/go.uuid"
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
			http.Error(w, "Internal server error", http.StatusBadRequest)
			return
		}

		//send data to be inserted into database & check for error
		u, err := signUserUp(un, bs, f, l)

		if err != nil {
			http.Error(w, "Uh oh. Something went wrong on our end. Try signing up again", http.StatusInternalServerError)
			return
		}

		//create a session cookie
		sID := uuid.NewV4().String()
		c := &http.Cookie{
			Name:  "session",
			Value: sID,
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

		u, err := logUserIn(un, p)

		if err != nil {
			http.Error(w, "user name and password do not match", http.StatusBadRequest)
			return
		}

		//create a cookie
		sID := uuid.NewV4().String()
		c := &http.Cookie{
			Name:  "session",
			Value: sID,
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
	//get user info from cookie session
	u := sessions.GetUser(req, Repo.DbUsers)

	if req.Method == http.MethodPost {
		//parse th form data
		req.ParseForm()

		//parse each field into []Workout
		wkout, err := logWorkout(req.Form["description"], req.Form["sets"], req.Form["reps"])

		if err != nil {
			http.Error(w, "workout not logged in bro!", http.StatusBadRequest)
		}
		log.Println(wkout)

		//need to create function to insert wkout into database with userid as foreign key
		//send workout info to be stored in database
		err = InsertGymSession(wkout, u.ID)

		if err != nil {
			http.Error(w, "Sorry. Unable to record gym session. Please try again", http.StatusBadRequest)
			return
		}

		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "gymsession.gohtml", nil)
}

func WorkoutEntry(w http.ResponseWriter, req *http.Request) {
	//Step 1: check to see if logged in
	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	var workout *models.Workout

	switch req.Method {
	case http.MethodGet:
		workout = readGymEntry(req, workout)
	case http.MethodPut:
		workout = updateGymEntry(req, workout)
	case http.MethodDelete:
		workout = deleteGymEntry(req, workout)
	default:
		http.Redirect(w, req, "/logbook", http.StatusSeeOther)
		return
	}
	w.WriteHeader(http.StatusAccepted)

	//Step 4: send data to template
	Repo.Template.ExecuteTemplate(w, "viewEntry.gohtml", workout)
}

func LogBook(w http.ResponseWriter, req *http.Request) {
	//Step 1: check to see if logged in
	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	//Step 2: get id from cookie value
	user := sessions.GetUser(req, Repo.DbUsers)
	//Step 3: search database for all workouts
	query := `select * from workouts where userid=$1`

	results, err := Repo.DB.Query(query, user.ID)
	if err != nil {
		http.Error(w, "Sorry. We are experiencing issues", http.StatusInternalServerError)
		return
	}

	var gymSession []models.GymSession

	for results.Next() {
		var wkout models.GymSession

		if err := results.Scan(&wkout.ID, &wkout.Workout, &wkout.UserID); err != nil {
			log.Fatalln(err)
		}
		gymSession = append(gymSession, wkout)
	}
	//Step 4: send data to template
	Repo.Template.ExecuteTemplate(w, "logbook.gohtml", gymSession)
}
