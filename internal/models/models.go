package models

import (
	"time"
)

//used to store user info in database and to authenticate user by accessing password info
type User struct {
	ID       string
	UserName string `json:"username"`
	Password []byte `json:"password"`
	First    string `json:"firstname"`
	Last     string `json:"lastname"`
}

//this data sent to template
type UserInfo struct {
	ID       string
	UserName string
	First    string
	Last     string
}

//each exercise in the workout will be held in this data structure
type Workout struct {
	Description string `json:"description"`
	Sets        int64  `json:"sets"`
	Reps        int64  `json:"reps"`
	Weight      int64  `json:"weight"`
}

//this will be used to hold each gym sessions workout info and send/receive to/from database
type GymSession struct {
	ID      string    `json:"id"`
	Workout []byte    `json:"workout"`
	UserID  string    `json:"userID"`
	Date    time.Time `json:"date"`
}

type GymLog struct {
	ID      string             `json:"id"`
	Workout map[string]Workout `json:"workout"`
	UserID  string             `json:"userID"`
	Date    string             `json:"date"`
}

type Chart struct {
	Label      string
	Labels     []string
	DataPoints []int64
}

type Data struct {
	Data         interface{}
	Keys         []string
	Count        int64
	User         UserInfo
	ErrorMessage string
	ChartData    Chart
}

//used to collect info from updating workout info
type FormData struct {
	Description []string `json:"description"`
	Sets        []string `json:"sets"`
	Reps        []string `json:"reps"`
	Weight      []string `json:"weight"`
}
