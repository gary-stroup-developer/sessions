package sessions

import (
	"gary-stroup-developer/sessions/internal/models"
	"net/http"
)

var DbUsers = map[string]models.User{} // user ID, user
var DbSessions = map[string]string{}   // session ID, user ID

func GetUser(req *http.Request, u map[string]models.User) models.User {
	c, err := req.Cookie("session")

	if err != nil {
		return models.User{}
	}

	return u[c.Value]
}

func AlreadyLoggedIn(req *http.Request, u map[string]models.User) bool {
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}

	user := u[c.Value]
	return user.UserName != ""

}
