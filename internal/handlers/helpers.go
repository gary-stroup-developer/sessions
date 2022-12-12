package handlers

import (
	"fmt"
	"gary-stroup-developer/sessions/internal/models"
	"log"
	"net/http"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
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

func logWorkout(desc []string, sets []string, reps []string, weight []string) ([]models.Workout, error) {
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

		//convert the string data from post request to int64 & check for error
		w, err := strconv.ParseInt(weight[i], 10, 0)
		if err != nil {
			return m, err
		}

		//no errors then the workout can be populated
		m = append(m, models.Workout{
			Description: desc[i],
			Sets:        s,
			Reps:        r,
			Weight:      w,
		})
		i++
	}
	return m, nil
}

func InsertGymSession(wo []byte, userid string) error {
	query := `INSERT into workouts (id, workout, userid, date) VALUES ($1, $2, $3, $4)`
	sessionID := uuid.NewV4().String()

	year := strconv.Itoa(time.Now().UTC().Year())
	month := time.Now().UTC().Month().String()
	day := strconv.Itoa(time.Now().UTC().Day())

	date := fmt.Sprintf(`%s-%s-%s`, year, month, day)
	//date := now.Format("2006-01-01")
	_, err := Repo.DB.Exec(query, sessionID, wo, userid, date)
	if err != nil {
		return err
	}
	return nil
}

func readGymEntry(r *http.Request, s string) models.GymSession {

	var wo models.GymSession
	//Step 3: search database for workout with that id
	query := `select * from workouts where id=$1`

	//need to create gymSession variable

	Repo.DB.QueryRow(query, s).Scan(&wo)

	return wo
}

func totalWorkoutCount(id string) int64 {

	var count int64
	//Step 3: search database for workout with that id
	query := `select count(*) from workouts where id=$1 and date_part('year', date) = date_part('year', CURRENT_DATE);`

	//need to create gymSession variable

	result, _ := Repo.DB.Query(query, id)

	defer result.Close()

	for result.Next() {
		result.Scan(&count)
	}
	return count
}

func getExerciseByNameData(id, name string) []int64 {
	var workouts []models.Workout
	var data []int64
	query := `select workout from workouts where id=$1 and date_part('year', date) = date_part('year', CURRENT_DATE) ORDER BY date;`

	results, err := Repo.DB.Query(query, id)

	if err != nil {
		log.Fatalln("could not get exercise by name")
	}
	defer results.Close()

	for results.Next() {
		results.Scan(&workouts)
	}

	for _, wkout := range workouts {
		if wkout.Description == name {
			data = append(data, wkout.Weight)
		}
	}
	return data
}

func updateGymEntry(r *http.Request, gym models.GymSession) error {
	//Step 3: search database for gymSession with that id
	//query := `select * from workouts where id=$1`

	query := `UPDATE workouts
		SET workout = $1
		WHERE id = $2;`

	_, err := Repo.DB.Exec(query, gym.Workout, gym.ID)

	if err != nil {
		return err
	}

	return nil
}

func deleteGymEntry(r *http.Request, s string) error {

	//Step 3: search database for workout with that id
	query := `DELETE from workouts WHERE id = $1;`

	_, err := Repo.DB.Exec(query, s)

	if err != nil {
		return err
	}

	return nil
}
