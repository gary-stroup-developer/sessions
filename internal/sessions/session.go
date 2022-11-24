package sessions

import (
	"gary-stroup-developer/sessions/internal/models"
	"net/http"
)

func GetUser(req *http.Request, u map[string]models.UserInfo) models.UserInfo {
	c, err := req.Cookie("session")

	if err != nil {
		return models.UserInfo{}
	}

	return u[c.Value]

}

func AlreadyLoggedIn(req *http.Request, u map[string]models.UserInfo) bool {
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}

	user := u[c.Value]
	return user.UserName != ""

}
