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
