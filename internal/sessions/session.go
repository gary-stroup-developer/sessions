package sessions

import (
	"gary-stroup-developer/sessions/internal/models"
	"net/http"
)

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
