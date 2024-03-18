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

func (h *HandlerController) createFilm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST request.")
	var f Film
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(f)
	err = h.repo.createFilm(&f)
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
	var actorsWithFilm []ActorWithFilms
	for _, actor := range actors {
		var actorWithFilm ActorWithFilms
		actorWithFilm.Id = actor.Id
		actorWithFilm.Name = actor.Name
		actorWithFilm.Sex = actor.Sex
		actorWithFilm.Birthday = actor.Birthday
		actorWithFilm.Films = h.repo.getFilmsByActorId(actor.Id)
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

func (r *PostgresRepository) getFilmById(filmId int) Film {
	row := r.db.QueryRow(context.Background(), "select * from public.film where film_id=$1", filmId)

	var f Film
	var releaseDate time.Time
	err := row.Scan(&f.Id, &f.Name, &f.Description, &releaseDate, &f.Rating, &f.ActorList)
	f.ReleaseDate = fmt.Sprintf("%d-%02d-%02d", releaseDate.Year(), releaseDate.Month(), releaseDate.Day())
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func (r *PostgresRepository) getFilmsByActorId(actorId int) []Film {
	rows, err := r.db.Query(context.Background(), "select film_id from public.actor_film where actor_id=$1", actorId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var filmsId []int
	for rows.Next() {
		var filmId int
		err := rows.Scan(&filmId)
		if err != nil {
			log.Fatal(err)
		}
		filmsId = append(filmsId, filmId)
	}
	var films []Film
	for _, film_id := range filmsId {
		film := r.getFilmById(film_id)
		films = append(films, film)
	}

	return films
}

func (r *PostgresRepository) createFilm(f *Film) error {
	//  'Green card', 'asdjsd;f', '2017-03-20', 5, '{1, 2, 3}'
	query := fmt.Sprintf("insert into public.film(film_name, film_description, film_release_date, film_rating, film_actor_list) values($1, $2, $3, $4, $5) RETURNING film_id;")
	row := r.db.QueryRow(context.Background(), query, f.Name, f.Description, f.ReleaseDate, f.Rating, f.ActorList)
	var film_id int
	err := row.Scan(&film_id)
	if err != nil {
		log.Fatal(err)
	}
	for _, actor_id := range f.ActorList {
		commandTag, err := r.db.Exec(context.Background(), "INSERT INTO public.actor_film(actor_id, film_id) VALUES ($1, $2);", actor_id, film_id)
		if err != nil {
			log.Fatal(err)
		}
		if commandTag.RowsAffected() != 1 {
			log.Fatal(errors.New("Film->Actor not created"))
		}
	}
	return err
}

func (r *PostgresRepository) updateActor(actorId int, updates map[string]string) error {
	var parameters []string

	for k, v := range updates {
		parameters = append(parameters, fmt.Sprintf("%s='%s'", k, v))
	}

	query := fmt.Sprintf("UPDATE public.actor SET %s WHERE actor_id=%d;", strings.Join(parameters, ", "), actorId)
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

	mux := http.NewServeMux()

	mux.HandleFunc("GET /actor", handler.getAllActors)
	mux.HandleFunc("POST /actor", handler.createActor)
	mux.HandleFunc("PUT /actor", handler.updateActor)
	mux.HandleFunc("POST /film", handler.createFilm)

	err = http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
	}
}
