package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type Actor struct {
	Id       int    `json:"actor_id"`
	Name     string `json:"actor_name"`
	Sex      string `json:"actor_sex"`
	Birthday string `json:"actor_birthday"`
}

type ActorWithFilms struct {
	Id       int    `json:"actor_id"`
	Name     string `json:"actor_name"`
	Sex      string `json:"actor_sex"`
	Birthday string `json:"actor_birthday"`
	Films    []Film `json:"actor_films"`
}

type Film struct {
	Id          int    `json:"film_id"`
	Name        string `json:"film_name"`
	Description string `json:"film_description"`
	ReleaseDate string `json:"film_release_date"`
	Rating      int    `json:"film_rating"`
	ActorList   []int  `json:"film_actor_list"` // Во время вывода передавать имена актеров, а не их id
}

type PostgresRepository struct {
	db *pgx.Conn
}

type HandlerController struct {
	repo *PostgresRepository
}

func NewHandlerController(repo *PostgresRepository) *HandlerController {
	return &HandlerController{repo: repo}
}

func (h *HandlerController) createActor(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST request.")
	var a Actor
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.repo.createActor(&a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func (h *HandlerController) updateActor(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PUT request.")
	var a Actor
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(a)
	updates := make(map[string]string)
	if a.Id == 0 {
		err := errors.New("actor_id is required field.")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if a.Name != "" {
		updates["actor_name"] = a.Name
	}
	if a.Sex != "" {
		updates["actor_sex"] = a.Sex
	}
	if a.Birthday != "" {
		updates["actor_birthday"] = a.Birthday
	}
	err = h.repo.updateActor(a.Id, updates)
	if err != nil {
		fmt.Println("alksdjfalksjdf")
	}
}

func (h *HandlerController) getAllActors(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET request.")
	w.Header().Set("Content-Type", "application/json")
	actors := h.repo.getAllActors()
	// filmsByActorId := h.repo.getFilmsByActorId()
	var actorsWithFilm []ActorWithFilms
	for _, actor := range actors {
		var actorWithFilm ActorWithFilms
		actorWithFilm.Id = actor.Id
		actorWithFilm.Name = actor.Name
		actorWithFilm.Sex = actor.Sex
		actorWithFilm.Birthday = actor.Birthday
		actorsWithFilm = append(actorsWithFilm, actorWithFilm)
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(actorsWithFilm)
}

func NewPostgresRepository(db *pgx.Conn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) createActor(m *Actor) error {
	commandTag, err := r.db.Exec(context.Background(), "insert into public.actor(actor_name, actor_sex, actor_birthday) values($1, $2, $3);", m.Name, m.Sex, m.Birthday)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aaksdjf %v\n", err)
	}
	if commandTag.RowsAffected() != 1 {
		err := errors.New("Actor is not create.")
		return err
	}
	return err
}

func (r *PostgresRepository) getAllActors() []Actor {
	rows, err := r.db.Query(context.Background(), "select * from public.actor")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var actors []Actor
	for rows.Next() {
		var a Actor
		var birthday time.Time
		err := rows.Scan(&a.Id, &a.Name, &a.Sex, &birthday)
		if err != nil {
			log.Fatal(err)
		}
		a.Birthday = fmt.Sprintf("%d-%02d-%02d", birthday.Year(), birthday.Month(), birthday.Day())
		actors = append(actors, a)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return actors
}
func (r *PostgresRepository) getFilmsByActorId() []Film {
	var films []Film
	return films
}

func (r *PostgresRepository) updateActor(id int, updates map[string]string) error {
	var parameters []string

	for k, v := range updates {
		parameters = append(parameters, fmt.Sprintf("%s='%s'", k, v))
	}

	query := fmt.Sprintf("UPDATE public.actor SET %s WHERE actor_id=%d;", strings.Join(parameters, ", "), id)
	commandTag, err := r.db.Exec(context.Background(), query)

	if err != nil {
		fmt.Fprintf(os.Stderr, "aaksdjf %v\n", err)
	}

	if commandTag.RowsAffected() != 1 {
		err := errors.New("Actor is not update.")
		return err
	}
	return err
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
	postgresRepo := NewPostgresRepository(conn)
	handler := NewHandlerController(postgresRepo)

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

	mux := http.NewServeMux()

	mux.HandleFunc("GET /actor", handler.getAllActors)
	mux.HandleFunc("POST /actor", handler.createActor)
	mux.HandleFunc("PUT /actor", handler.updateActor)

	err = http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
	}
}
