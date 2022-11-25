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

type FormData map[string][]string

type Workout struct {
	Description string
	Sets        int64
	Reps        int64
}
