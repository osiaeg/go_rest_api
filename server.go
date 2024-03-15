package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type Actor struct {
	Name     string `json:"actor_name"`
	Sex      string `json:"actor_sex"`
	Birthday string `json:"actor_birthday"`
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	// Add logger
	if r.Method == "GET" {
		fmt.Println("GET request.")
	} else if r.Method == "POST" {
		fmt.Println("POST request.")
		var a Actor
		err := json.NewDecoder(r.Body).Decode(&a)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	io.WriteString(w, "Hello, HTTP! \n")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file.")
	}

	// urlExample := "postgres://user:admin@localhost:54320/postgres"
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// Insert Actor
	//

	// Insert Film
	/*
		insert into public.film(
			film_name,
			film_description,
			film_release_date,
			film_rating,
			film_actor_list
			)
			values(
			'Green card',
			'asdjf;lakjsdl;kfjal;skdjflfjaskl;djf;laskdjf;kljasdkl;fjasd;f',
			'2017-03-20',
			5,
			'{1, 2, 3}'
		);
	*/

	commandTag, err := conn.Exec(context.Background(), "insert into public.actor(actor_name, actor_sex, actor_birthday) values($1, $2, $3);", "Jone2", "F", "2017-03-14")
	if err != nil {
		fmt.Fprintf(os.Stderr, "aaksdjf %v\n", err)
	}
	if commandTag.RowsAffected() != 1 {
		fmt.Println("Something goes wrong.")
	}

	http.HandleFunc("/actor", getRoot)
	http.HandleFunc("/hello", getHello)

	err = http.ListenAndServe(":8080", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
	}
}
