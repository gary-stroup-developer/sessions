package handlers

import (
	"gary-stroup-developer/sessions/internal/models"
	"net/http"
	"strconv"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

//function will insert user info into db and return the UserInfo{}
func signUserUp(un string, bs []byte, f string, l string) (models.UserInfo, error) {
	id := uuid.NewV4().String()
	sqlStatement := `
		insert into users (id,username, password, firstname, lastname)
		values ($1, $2, $3, $4, $5)
		returning id,username,firstname,lastname`

	user := models.UserInfo{}

	err := Repo.DB.QueryRow(sqlStatement, id, un, bs, f, l).Scan(&user.ID, &user.UserName, &user.First, &user.Last)
	if err != nil {
		return user, err
	}

	return user, nil
}

//function will log user in if found in database and passwords match
func logUserIn(un string, p string) (models.UserInfo, error) {
	//initialize user struct and individual fields that will accept values from query result
	var u models.User

	//make a request to get user info from DB
	err := Repo.DB.QueryRow(`select * from users where username=$1`, un).Scan(&u.ID, &u.UserName, &u.Password, &u.First, &u.Last)
	if err != nil {
		return models.UserInfo{}, err
	}

	//check if returned password matches the password submitted by form
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p))

	if err != nil {
		return models.UserInfo{}, err
	}

	return models.UserInfo{ID: u.ID, UserName: u.UserName, First: u.First, Last: u.Last}, nil
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

func InsertGymSession(wo []models.Workout, userid string) error {
	query := `INSERT into workouts (id, workout, userid) VALUES ($1, $2, $3)`
	sessionID := uuid.NewV4().String()
	_, err := Repo.DB.Exec(query, sessionID, wo, userid)
	if err != nil {
		return err
	}
	return nil
}

func readGymEntry(r *http.Request, wo *models.Workout) *models.Workout {
	//Step 2: get id from url
	// id, _ := url.Parse("http://localhost:8080/workout/?id=55")
	userID := r.URL.Query()["id"][0]

	//Step 3: search database for workout with that id
	query := `select * from workouts where id=$1`

	//need to create gymSession variable

	Repo.DB.QueryRow(query, userID).Scan(&wo)

	return wo
}

func updateGymEntry(r *http.Request, wo *models.Workout) *models.Workout {
	//Step 2: get id from url
	// id, _ := url.Parse("http://localhost:8080/workout/?id=55")
	userID := r.URL.Query()["id"][0]

	//Step 3: search database for gymSession with that id
	query := `select * from workouts where id=$1`

	Repo.DB.QueryRow(query, userID).Scan(&wo)

	return wo
}

func deleteGymEntry(r *http.Request, wo *models.Workout) *models.Workout {
	//Step 2: get id from url
	// id, _ := url.Parse("http://localhost:8080/workout/?id=55")
	userID := r.URL.Query()["id"][0]

	//Step 3: search database for workout with that id
	query := `UPDATE workouts
		SET workout = $1,
    	sets = $2,
		reps = $3,
		notes = $4,
		WHERE id = $1;`

	Repo.DB.QueryRow(query, userID).Scan(&wo)

	return wo
}
