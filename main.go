package main

import (
	"fmt"
	"gary-stroup-developer/sessions/internal/driver"
	"gary-stroup-developer/sessions/internal/handlers"
	"gary-stroup-developer/sessions/internal/routes"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("./internal/templates/*"))
}

func main() {

	godotenv.Load(".env")

	var pass = os.Getenv("PASSWORD")
	var user = os.Getenv("USER")
	var dbname = os.Getenv("DB")

	port, ok := os.LookupEnv("PORT")

	if !ok {
		port = "80"
	}

	var postgresqlDbInfo = fmt.Sprintf("host=localhost port=5432 user=%s "+
		"password=%s dbname=%s",
		user, pass, dbname)

	db := driver.ConnectDB(postgresqlDbInfo)

	defer db.Close()

	server := http.NewServeMux()

	repo := handlers.NewRepo(db, tpl)
	handlers.SetRepo(repo)
	routes.Routes(server)

	log.Fatal(http.ListenAndServe(":"+port, server))
}
