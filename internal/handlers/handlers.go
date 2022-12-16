package handlers

import (
	"database/sql"
	"encoding/json"
	"gary-stroup-developer/sessions/internal/models"
	"gary-stroup-developer/sessions/internal/sessions"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	var data models.Data

	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	u := sessions.GetUser(req, Repo.DbUsers)
	data.User = u

	var chartPoints = make([]int64, 12)
	var labels []string

	totalCount := totalWorkoutCount(u.ID)

	if totalCount == 0 {
		data.ChartData.DataPoints = append(chartPoints, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
		data.ChartData.Labels = append(labels, "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec")

	}
	data.Count = totalCount

	if req.Method == http.MethodPost {
		exercise := req.PostForm.Get("exercise")

		exerciseData := getExerciseByNameData(u.ID, exercise)

		for i := 0; i < len(exerciseData); i++ {
			data.ChartData.Labels = append(data.ChartData.Labels, strconv.Itoa(i))
		}

		data.ChartData.DataPoints = exerciseData
		data.ChartData.Label = exercise

	}

	Repo.Template.ExecuteTemplate(w, "dashboard.gohtml", data)
}

func GetKeys(w http.ResponseWriter, req *http.Request) {
	year := strconv.Itoa(time.Now().UTC().Year())
	month := time.Now().UTC().Month().String()

	data := models.Data{}

	u := sessions.GetUser(req, Repo.DbUsers)
	id := u.ID

	var keys []string

	query := `SELECT json_object_keys (workout) as keys
			  FROM workouts
			  WHERE userid=$1 and extract(year from "date") = $2 and extract(MONTH from "date") = $3
			  group by keys;`

	results, err := Repo.DB.Query(query, id, year, month)

	if err != nil {
		http.Error(w, "something went wrong querying the DB", http.StatusBadRequest)
	}
	defer results.Close()

	for results.Next() {
		results.Scan(&keys)
	}

	data.Keys = keys

	w.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(data.Keys)
	w.Write(resp)
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

		//parse each field into []Workout
		wkout, err := logWorkout(req.Form["description"], req.Form["sets"], req.Form["reps"], req.Form["weight"])

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			data.ErrorMessage = "workout not logged in bro!"
			return
		}
		workout, _ := json.Marshal(&wkout)
		//need to create function to insert wkout into database with userid as foreign key
		//send workout info to be stored in database
		err = InsertGymSession(workout, u.ID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			data.ErrorMessage = "Sorry. Unable to record gym session. Please try again"
			return
		}

		err = json.Unmarshal(workout, &wkout)
		if err != nil {
			log.Println("cannot unmarshal workout")
		}
		log.Println(wkout)

		http.Redirect(w, req, "/dashboard", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "gymsession.gohtml", data)
}

func LogBook(w http.ResponseWriter, req *http.Request) {
	//Step 1: check to see if logged in
	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	//Step 2: get id from cookie value
	user := sessions.GetUser(req, Repo.DbUsers)
	//log.Println(user)
	//Step 3: search database for all workouts
	query := `SELECT id, workout, userid, "date"
		FROM workouts
		WHERE userid=$1`

	results, err := Repo.DB.Query(query, user.ID)
	if err != nil {
		http.Error(w, "Sorry. We are experiencing issues finding workout entry", http.StatusInternalServerError)
		return
	}
	defer results.Close()

	var gymSession []models.GymLog

	for results.Next() {
		var wkout models.GymSession
		var data models.GymLog

		if err := results.Scan(&wkout.ID, &wkout.Workout, &wkout.UserID, &wkout.Date); err != nil {
			http.Error(w, "could not retrieve workout entries", http.StatusInternalServerError)
			return
		}

		var workout map[string]models.Workout

		err = json.Unmarshal(wkout.Workout, &workout)
		if err != nil {
			log.Fatalln(err)
		}
		data.ID = wkout.ID
		data.Workout = workout
		data.UserID = wkout.UserID
		data.Date = strings.Split(wkout.Date.String(), " ")[0]

		gymSession = append(gymSession, data)

	}

	//Step 4: send data to template
	Repo.Template.ExecuteTemplate(w, "logbook.gohtml", gymSession)
}

func EditWorkoutEntry(w http.ResponseWriter, req *http.Request) {
	//Step 1: check to see if logged in
	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	var data models.FormData
	if req.Method == http.MethodPut {
		//Step 2: get id from url
		// id, _ := url.Parse("http://localhost:8080/workout/?id=55")
		gymID := req.URL.Query()["id"][0]
		values, _ := io.ReadAll(req.Body)
		err := json.Unmarshal(values, &data)
		if err != nil {
			log.Println("no body found")
		}

		// parse each field into []Workout
		workout, err := logWorkout(data.Description, data.Sets, data.Reps, data.Weight)

		if err != nil {
			http.Error(w, "workout updates was not inserted into the DB bro!", http.StatusBadRequest)
			return
		}
		wkout, err := json.Marshal(&workout)
		if err != nil {
			http.Error(w, "workout not updated bro!", http.StatusBadRequest)
			return
		}

		workoutSession := models.GymSession{ID: gymID, Workout: wkout}

		err = updateGymEntry(req, workoutSession)
		if err != nil {
			http.Error(w, "workout not updated bro!", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, _ := json.Marshal(map[string]string{"message": "workout has been updated!"})
		w.Write(jsonResp)
	}
}

func DeleteWorkoutEntry(w http.ResponseWriter, req *http.Request) {
	//Step 1: check to see if logged in
	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	//Step 2: get id from url
	// id, _ := url.Parse("http://localhost:8080/workout/?id=55")
	gymID := req.URL.Query()["id"][0]

	if req.Method == http.MethodDelete {
		err := deleteGymEntry(req, gymID)
		if err != nil {
			http.Error(w, "workout not deleted bro!", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, _ := json.Marshal(map[string]string{"message": "workout has been deleted!"})
		w.Write(jsonResp)
	}

}
