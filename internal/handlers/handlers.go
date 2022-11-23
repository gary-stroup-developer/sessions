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
)

type Repository struct {
	Template *template.Template
	DB       *sql.DB
}

func NewRepo(db *sql.DB, tpl *template.Template) *Repository {
	return &Repository{
		Template: tpl,
		DB:       db,
	}
}

var Repo *Repository

func SetRepo(r *Repository) {
	Repo = r
}

func Index(w http.ResponseWriter, req *http.Request) {
	u := sessions.GetUser(w, req)

	Repo.Template.ExecuteTemplate(w, "index.gohtml", u)
}

func Bar(w http.ResponseWriter, req *http.Request) {
	u := sessions.GetUser(w, req)

	if !sessions.AlreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "bar.gohtml", u)
}

func Signup(w http.ResponseWriter, req *http.Request) {
	if sessions.AlreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	var u models.User

	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")
		f := req.FormValue("firstname")
		l := req.FormValue("lastname")

		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}

		http.SetCookie(w, c)

		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)

		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		u = models.User{UserName: un, Password: bs, First: f, Last: l}

		//db := driver.ConnectDB()

		sqlStatement := `
		insert into users (username, password, firstname, lastname)
		values ($1, $2, $3, $4)
		returning firstname`
		fname := ""
		//db := driver.ConnectDB()
		err = Repo.DB.QueryRow(sqlStatement, un, bs, f, l).Scan(&fname)
		if err != nil {
			log.Println(err.Error())
		}
		// defer db.Close()
		fmt.Println("New record ID is:", fname)
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "signup.gohtml", u)
}

func Login(w http.ResponseWriter, req *http.Request) {
	if sessions.AlreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")

		u, ok := sessions.DbUsers[un]
		if !ok {
			http.Error(w, "username and/or password do not match", http.StatusForbidden)
			return
		}

		err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))

		if err != nil {
			http.Error(w, "username and/or password do not match", http.StatusForbidden)
			return
		}

		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		sessions.DbSessions[c.Value] = un
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	Repo.Template.ExecuteTemplate(w, "login.gohtml", nil)
}
