package models

type User struct {
	UserName string
	Password []byte
	First    string
	Last     string
}

type UserInfo struct {
	UserName string
	First    string
	Last     string
}

//this will be used to capture the workout submission
type FormData map[string][]string

//each exercise in the workout will be held in this data structure
type Workout struct {
	Description string
	Sets        int64
	Reps        int64
}
