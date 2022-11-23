package sessions

import (
	"gary-stroup-developer/sessions/internal/models"
	"net/http"

	"github.com/gofrs/uuid"
)

var DbUsers = map[string]models.User{} // user ID, user
var DbSessions = map[string]string{}   // session ID, user ID

func GetUser(w http.ResponseWriter, req *http.Request) models.User {
	c, err := req.Cookie("session")

	if err != nil {
		sID, _ := uuid.NewV4()
		c = &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
	}
	http.SetCookie(w, c)

	// if user exists already, get user
	var u models.User

	if un, ok := DbSessions[c.Value]; ok {
		u = DbUsers[un]
	}

	return u
}

func AlreadyLoggedIn(req *http.Request) bool {
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}

	un := DbSessions[c.Value]
	_, ok := DbUsers[un]
	return ok
}
