package handlers

import (
	"gary-stroup-developer/sessions/internal/models"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

//function will insert user info into db and return the UserInfo{}
func signUserUp(un string, bs []byte, f string, l string) (models.UserInfo, error) {
	sqlStatement := `
		insert into users (username, password, firstname, lastname)
		values ($1, $2, $3, $4)
		returning username,firstname,lastname`

	user := models.UserInfo{}

	err := Repo.DB.QueryRow(sqlStatement, un, bs, f, l).Scan(&user.UserName, &user.First, &user.Last)
	if err != nil {
		return user, err
	}

	return user, nil
}

//function will log user in if found in database and passwords match
func logUserIn(w http.ResponseWriter, un string, p string) (models.UserInfo, error) {
	//initialize user struct and individual fields that will accept values from query result
	var u models.User

	//make a request to get user info from DB
	err := Repo.DB.QueryRow(`select * from users where username=$1`, un).Scan(&u.UserName, &u.Password, &u.First, &u.Last)
	if err != nil {
		http.Error(w, "user not found", http.StatusBadRequest)
		return models.UserInfo{}, err
	}

	//check if returned password matches the password submitted by form
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p))

	if err != nil {
		http.Error(w, "username and/or password do not match", http.StatusForbidden)
		return models.UserInfo{}, err
	}

	return models.UserInfo{UserName: u.UserName, First: u.First, Last: u.Last}, nil
}

func logWorkout(desc []string, sets []string, reps []string) ([]models.Workout, error) {
	var i = 0
	var m []models.Workout

	for i < len(desc) {
		//convert the string data from post request to int64 & check for error
		s, err := strconv.ParseInt(sets[i], 10, 0)
		if err != nil {
			return m, err
		}
		//convert the string data from post request to int64 & check for error
		r, err := strconv.ParseInt(reps[i], 10, 0)
		if err != nil {
			return m, err
		}

		//no errors then the workout can be populated
		m = append(m, models.Workout{
			Description: desc[i],
			Sets:        s,
			Reps:        r,
		})
		i++
	}
	return m, nil
}

// func getAllEntries() {
// 	sqlStatement := `select * from users`
// 	rows, err := Repo.DB.Query(sqlStatement)

// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var (
// 			id   int64
// 			name string
// 		)
// 		if err := rows.Scan(&id, &name); err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Printf("id %d name is %s\n", id, name)
// 	}
// 	if !rows.NextResultSet() {
// 		log.Fatalf("expected more result sets: %v", rows.Err())
// 	}
// }
