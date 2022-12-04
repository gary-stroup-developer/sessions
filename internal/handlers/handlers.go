package handlers

import (
	"database/sql"
	"encoding/json"
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

//works fine
func Index(w http.ResponseWriter, req *http.Request) {

	Repo.Template.ExecuteTemplate(w, "index.gohtml", nil)
}

//works fine but needs to be completed
func Dashboard(w http.ResponseWriter, req *http.Request) {

	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	u := sessions.GetUser(req, Repo.DbUsers)

	Repo.Template.ExecuteTemplate(w, "dashboard.gohtml", u)
}

//works fine
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
			http.Error(w, "username and password don't match", http.StatusBadRequest)
			return
		}

		//send data to be inserted into database & check for error
		u, err := signUserUp(un, bs, f, l)

		if err != nil {
			http.Error(w, "username and password don't match", http.StatusBadRequest)
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

//works fine
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
			http.Error(w, "username and password don't match", http.StatusBadRequest)
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

//works fine
func GymSession(w http.ResponseWriter, req *http.Request) {
	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	//get user info from cookie session
	u := sessions.GetUser(req, Repo.DbUsers)

	var data models.Data

	if req.Method == http.MethodPost {
		//parse th form data
		req.ParseForm()
		log.Println(req.Form["description"], req.Form["sets"], req.Form["reps"], req.Form["weight"])
		//parse each field into []Workout
		wkout, err := logWorkout(req.Form["description"], req.Form["sets"], req.Form["reps"], req.Form["weight"])

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			data.ErrorMessage["message"] = "workout not logged in bro!"
			return
		}

		//need to create function to insert wkout into database with userid as foreign key
		//send workout info to be stored in database
		err = InsertGymSession(wkout, u.ID)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			data.ErrorMessage["message"] = "Sorry. Unable to record gym session. Please try again"
			return
		}

		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "gymsession.gohtml", data)
}

func WorkoutEntry(w http.ResponseWriter, req *http.Request) {
	//Step 1: check to see if logged in
	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	//Step 2: get id from url
	// id, _ := url.Parse("http://localhost:8080/workout/?id=55")
	gymID := req.URL.Query()["id"][0]

	var workoutSession models.GymSession
	var data models.Data

	switch req.Method {

	case http.MethodGet:
		if gymID == "" {
			w.WriteHeader(http.StatusBadRequest)
			data.ErrorMessage["message"] = "Unable to retreive gym entry"
			return
		}

		workoutSession = readGymEntry(req, gymID)
		data.Data = workoutSession
		//Step 4: send data to template
		Repo.Template.ExecuteTemplate(w, "viewEntry.gohtml", data)

	case http.MethodPut:
		req.ParseForm()

		//parse each field into []Workout
		workout, err := logWorkout(req.Form["description"], req.Form["sets"], req.Form["reps"], req.Form["weight"])

		if err != nil {
			http.Error(w, "workout not logged in bro!", http.StatusBadRequest)
			return
		}

		workoutSession := models.GymSession{ID: gymID, Workout: workout}

		err = updateGymEntry(req, workoutSession)
		if err != nil {
			http.Error(w, "workout not updated bro!", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, _ := json.Marshal(map[string]string{"message": "workout has been updated!"})
		w.Write(jsonResp)

	case http.MethodDelete:
		err := deleteGymEntry(req, gymID)
		if err != nil {
			http.Error(w, "workout not deleted bro!", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, _ := json.Marshal(map[string]string{"message": "workout has been deleted!"})
		w.Write(jsonResp)
	default:
		http.Redirect(w, req, "/logbook", http.StatusSeeOther)
		return
	}

}

func LogBook(w http.ResponseWriter, req *http.Request) {
	//Step 1: check to see if logged in
	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	//Step 2: get id from cookie value
	user := sessions.GetUser(req, Repo.DbUsers)
	log.Println(user)
	//Step 3: search database for all workouts
	query := `SELECT id, userid, "date"
		FROM workouts
		WHERE userid=$1`

	results, err := Repo.DB.Query(query, user.ID)
	if err != nil {
		http.Error(w, "Sorry. We are experiencing issues finding workout entry", http.StatusInternalServerError)
		return
	}
	defer results.Close()
	var gymSession []models.GymSession

	for results.Next() {
		var wkout models.GymSession

		if err := results.Scan(&wkout.ID, &wkout.UserID, &wkout.Date); err != nil {
			http.Error(w, "could not retrieve workout entries", http.StatusInternalServerError)
			return
		}
		gymSession = append(gymSession, wkout)
	}
	//Step 4: send data to template
	Repo.Template.ExecuteTemplate(w, "logbook.gohtml", gymSession)
}
