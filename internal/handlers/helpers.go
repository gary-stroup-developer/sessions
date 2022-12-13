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

// **********************************************************************************************************************************
// SELECT workout, extract(year from "date") as YEAR,
// extract(MONTH from "date") as MONTH,
// extract(DAY from "date") as DAY
// FROM public.workouts
// where userid = $1 ='15fdfe2f-177f-4ff7-b97b-61d7e691ea3d'
// order by month,day;

// func getChartData(id string, name string) []int64 {
// 	var workouts []models.Workout

// 	var year string
// 	var month string
// 	var day string

// 	var data []int64

// 	monthlyQuery := `SELECT workout, extract(year from "date") as YEAR,
// 			  extract(MONTH from "date") as MONTH,
// 			  extract(DAY from "date") as DAY
// 			  FROM public.workouts
// 			  where userid = $1 and extract(MONTH from "date") = $2 and extract(year from "date") = $3
// 			  order by month,day;`

// 	annualQuery := `SELECT workout, extract(year from "date") as YEAR,
// 			  extract(MONTH from "date") as MONTH,
// 			  extract(DAY from "date") as DAY
// 			  FROM public.workouts
// 			  where userid = $1 and extract(year from "date") = $2
// 			  order by month,day;`
// 	weeklyQuery := `SELECT workout, extract(year from "date") as YEAR,
// 			  extract(MONTH from "date") as MONTH,
// 			  extract(DAY from "date") as DAY
// 			  FROM public.workouts
// 			  where userid = $1 and extract(MONTH from "date") = $2 and extract(year from "date") = $3
// 			  and extract(DAY from "date") <= $4 and extract(DAY from "date") >= $5
// 			  order by day;`

// 	results, err := Repo.DB.Query(query, id)

// 	if err != nil {
// 		log.Fatalln("could not get exercise by name")
// 	}
// 	defer results.Close()

// 	for results.Next() {
// 		results.Scan(&workouts, &year, &month, &day)
// 	}

// 	for _, wkout := range workouts {
// 		if wkout.Description == name {
// 			data = append(data, wkout.Weight)
// 		}
// 	}
// 	return data
// }

// workout                                                                                                                                                                       |year|month|day|
// ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+----+-----+---+
// [{"description":"Leg Press","sets":5,"reps":5,"weight":360},{"description":"Squats","sets":5,"reps":5,"weight":225},{"description":"deadlift","sets":5,"reps":5,"weight":405}]|2022|    9|  1|
// [{"description":"Leg Press","sets":5,"reps":5,"weight":360},{"description":"Squats","sets":5,"reps":5,"weight":225},{"description":"deadlift","sets":5,"reps":5,"weight":405}]|2022|   10| 12|
// [{"description":"Leg Press","sets":5,"reps":5,"weight":360},{"description":"Squats","sets":5,"reps":5,"weight":225},{"description":"deadlift","sets":5,"reps":5,"weight":405}]|2022|   10| 25|
// [{"description":"Leg Press","sets":5,"reps":5,"weight":360},{"description":"Squats","sets":5,"reps":5,"weight":225},{"description":"deadlift","sets":5,"reps":5,"weight":405}]|2022|   11| 15|
// [{"description":"Leg Press","sets":5,"reps":5,"weight":360},{"description":"Squats","sets":5,"reps":5,"weight":225},{"description":"deadlift","sets":5,"reps":5,"weight":405}]|2022|   12|  7|
// [{"description":"Leg Press","sets":5,"reps":5,"weight":360},{"description":"Squats","sets":5,"reps":5,"weight":225},{"description":"deadlift","sets":5,"reps":5,"weight":405}]|2022|   12| 10|
// [{"description":"deadlift","sets":5,"reps":5,"weight":375},{"description":"Pull-ups","sets":5,"reps":10,"weight":0},{"description":"OHP","sets":8,"reps":8,"weight":185}]     |2022|   12| 12|
// ***********************************************************************************************************************************

// `SELECT workout, extract(year from "date") as YEAR,
// 			  extract(MONTH from "date") as MONTH,
// 			  extract(DAY from "date") as DAY
// 			  FROM public.workouts
// 			  where userid = $1 and extract(MONTH from "date") = $2
// 			  order by month,day;`

// workout                                                                                                                                                                       |year|month|day|
// ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+----+-----+---+
// [{"description":"Leg Press","sets":5,"reps":5,"weight":360},{"description":"Squats","sets":5,"reps":5,"weight":225},{"description":"deadlift","sets":5,"reps":5,"weight":405}]|2022|   10| 12|
// [{"description":"Leg Press","sets":5,"reps":5,"weight":360},{"description":"Squats","sets":5,"reps":5,"weight":225},{"description":"deadlift","sets":5,"reps":5,"weight":405}]|2022|   10| 25|
