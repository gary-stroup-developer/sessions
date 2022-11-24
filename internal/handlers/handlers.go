package handlers

import (
	"database/sql"
	"fmt"
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
	DbUsers  map[string]models.User // session ID, user
}

func NewRepo(db *sql.DB, tpl *template.Template) *Repository {
	return &Repository{
		Template: tpl,
		DB:       db,
		DbUsers:  make(map[string]models.User),
	}
}

var Repo *Repository

func SetRepo(r *Repository) {
	Repo = r
}

func Index(w http.ResponseWriter, req *http.Request) {

	Repo.Template.ExecuteTemplate(w, "index.gohtml", nil)
}

func Bar(w http.ResponseWriter, req *http.Request) {

	if !sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	u := sessions.GetUser(req, Repo.DbUsers)

	Repo.Template.ExecuteTemplate(w, "bar.gohtml", u)
}

func Signup(w http.ResponseWriter, req *http.Request) {
	if sessions.AlreadyLoggedIn(req, Repo.DbUsers) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	//initialize a user variable to store user info received from DB query
	var u models.User

	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")
		f := req.FormValue("firstname")
		l := req.FormValue("lastname")

		//create a session cookie
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		//store cookie in browser
		http.SetCookie(w, c)

		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		sqlStatement := `
		insert into users (username, password, firstname, lastname)
		values ($1, $2, $3, $4)
		returning firstname`
		fname := ""

		err = Repo.DB.QueryRow(sqlStatement, un, bs, f, l).Scan(&fname)
		if err != nil {
			log.Println(err.Error())
		}

		//store form data in user struct since we know password encryption & DB exection were successful
		u = models.User{UserName: un, Password: bs, First: f, Last: l}
		//bind session cookie with user
		Repo.DbUsers[c.Value] = u

		fmt.Println("New record ID is:", fname)
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "signup.gohtml", u)
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

		//initialize user struct and individual fields that will accept values from query result
		var u models.User
		var username string
		var pass string
		var first string
		var last string

		//make a request to get user info from DB
		err := Repo.DB.QueryRow(`select * from users where username=$1`, un).Scan(&username, &pass, &first, &last)
		if err != nil {
			http.Error(w, "user not found", http.StatusBadRequest)
			return
		}

		//store query result into user struct u
		u.UserName = username
		u.Password = []byte(pass)
		u.First = first
		u.Last = last

		//check if returned password matches the password submitted by form
		err = bcrypt.CompareHashAndPassword(u.Password, []byte(p))

		if err != nil {
			http.Error(w, "username and/or password do not match", http.StatusForbidden)
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

		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "login.gohtml", nil)
}
