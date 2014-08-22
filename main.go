package main

import (
	"encoding/json"
	"fmt"
	//two templates:
	"html/template"
	//also text/template
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Create a type alias
// Normally this would just be called "Sex"
// The T (for "type") is added here just for clarity
type SexT int

const (
	Unknown SexT = iota
	Male
	Female
)

type Person struct {
	Name string
	Sex  SexT
}

func (p Person) Greet() string {
	switch p.Name {
	case "Alice":
		return "Hey, Bob!"
	case "Bob":
		return "Hey, Alice!"
	}
	return "Eve is eavesdropping!"
}

var alice Person = Person{"Alice", Female}
var bob = Person{"Bob", Male}

//safeHandler is kind of like a function decorator since it's
//a function that accepts functions as arguments and returns a new
//function!
type safeHandler func(w http.ResponseWriter, r *http.Request) error

func (h safeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		handleError(w, r, err)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("ERROR: Recovern from panic: %v", err)
	http.Error(w, "An unexpected server error has occurred", http.StatusInternalServerError)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	//Declare new template, must pass name
	t := template.New("base")

	//parseFiles can also take multiple arguments
	s1, err := t.ParseFiles("templates/base.tmpl")
	if err != nil {
		// TODO don't panic!
		panic(err)
	}

	err = s1.ExecuteTemplate(w, "base", alice)
	if err != nil {
		// TODO don't panic!
		panic(err)
	}
}

func serveUserJson(w http.ResponseWriter, r *http.Request) error {
	var user Person
	vars := mux.Vars(r)
	switch vars["username"] {
	case "alice":
		user = alice
	case "bob":
		user = bob
	default:
		return fmt.Errorf("User not supported")
	}

	bts, err := json.Marshal(user)
	if err != nil {
		return err
	}
	//Add http header to response object w
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(bts)
	return nil
}

func serveChocolates(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	number_s := vars["number"]
	n, err := strconv.ParseInt(number_s, 10, 64)
	if err != nil {
		return err
	}

	t := template.New("base")
	s1, err := t.ParseFiles("templates/chocolates.tmpl")
	err = s1.ExecuteTemplate(w, "base", n)
	return err

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", serveHome)
	r.Handle("/users/{username:[A-Za-z0-9]+}.json", safeHandler(serveUserJson))
	r.Handle("/chocolates/{username:[A-Za-z0-9]+}/{number}", safeHandler(serveChocolates))
	http.Handle("/", r)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Error listening, %v", err)
	}
}
