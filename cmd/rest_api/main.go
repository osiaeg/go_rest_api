package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

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
func (h *HandlerController) updateFilm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PUT request.")
	var f Film
	f.Rating = -1
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updates := make(map[string]string)
	if f.Id == 0 {
		err := errors.New("film_id is required field.")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if f.Name != "" {
		updates["film_name"] = f.Name
	}
	if f.Description != "" {
		updates["film_description"] = f.Description
	}
	if f.ReleaseDate != "" {
		updates["film_release_date"] = f.ReleaseDate
	}
	if f.Rating != -1 {
		updates["film_rating"] = fmt.Sprintf("%d", f.Rating)
	}
	if len(f.ActorList) > 0 {
		var actors_id []string

		for _, actor_id := range f.ActorList {
			actors_id = append(actors_id, fmt.Sprintf("%d", actor_id))
		}

		updates["film_actor_list"] = fmt.Sprintf("{%s}", strings.Join(actors_id, ", "))
	}
	err = h.repo.updateFilm(f.Id, updates, f.ActorList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

func (h *HandlerController) deleteActor(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DELETE request.")
	actor_id := r.PathValue("id")
	err := h.repo.deleteActor(actor_id)
	if err != nil {
		log.Fatal(err)
	}
}
func (h *HandlerController) deleteFilm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DELETE request.")
	film_id := r.PathValue("id")
	err := h.repo.deleteFilm(film_id)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *HandlerController) getSortedFilms(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET request.")
	field_name := r.PathValue("field_name")
	order := r.PathValue("order")

	w.Header().Set("Content-Type", "application/json")
	films := h.repo.getSortedFilms(field_name, order)
	json.NewEncoder(w).Encode(films)
}

func (h *HandlerController) getFilms(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET request.")
	w.Header().Set("Content-Type", "application/json")
	films := h.repo.getSortedFilms("film_rating", "desc")
	json.NewEncoder(w).Encode(films)
}
func (h *HandlerController) getFilmsByName(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET request.")
	w.Header().Set("Content-Type", "application/json")
	part_of_name := r.PathValue("part_of_name")
	films := h.repo.searchFilmByName(part_of_name)
	json.NewEncoder(w).Encode(films)
}

func NewPostgresRepository(db *pgx.Conn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) searchFilmByName(part_of_name string) []Film {
	query := fmt.Sprintf("select * from public.film where film_name like '%%%s%%'", part_of_name)
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var films []Film
	for rows.Next() {
		var f Film
		var releaseDate time.Time
		err := rows.Scan(&f.Id, &f.Name, &f.Description, &releaseDate, &f.Rating, &f.ActorList)
		if err != nil {
			log.Fatal(err)
		}
		f.ReleaseDate = fmt.Sprintf("%d-%02d-%02d", releaseDate.Year(), releaseDate.Month(), releaseDate.Day())
		films = append(films, f)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return films
}

func (r *PostgresRepository) deleteActor(actor_id string) error {
	commandTag, err := r.db.Exec(context.Background(), "delete from public.actor where actor_id=$1", actor_id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aaksdjf %v\n", err)
	}
	if commandTag.RowsAffected() != 1 {
		err := errors.New("Actor is not deleted.")
		return err
	}
	return err
}

func (r *PostgresRepository) deleteFilm(film_id string) error {
	commandTag, err := r.db.Exec(context.Background(), "delete from public.film where film_id=$1", film_id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aaksdjf %v\n", err)
	}
	if commandTag.RowsAffected() != 1 {
		err := errors.New("Film is not deleted.")
		return err
	}
	return err
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

func (r *PostgresRepository) getSortedFilms(field_name string, order string) []Film {
	query := fmt.Sprintf("select * from public.film order by %s %s", field_name, order)
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var films []Film
	for rows.Next() {
		var f Film
		var releaseDate time.Time
		err := rows.Scan(&f.Id, &f.Name, &f.Description, &releaseDate, &f.Rating, &f.ActorList)
		if err != nil {
			log.Fatal(err)
		}
		f.ReleaseDate = fmt.Sprintf("%d-%02d-%02d", releaseDate.Year(), releaseDate.Month(), releaseDate.Day())
		films = append(films, f)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return films
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
			return err
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

func (r *PostgresRepository) updateFilm(filmId int, updates map[string]string, actorList []int) error {
	var parameters []string

	for k, v := range updates {
		parameters = append(parameters, fmt.Sprintf("%s='%s'", k, v))
	}

	query := fmt.Sprintf("UPDATE public.film SET %s WHERE film_id=%d;", strings.Join(parameters, ", "), filmId)
	commandTag, err := r.db.Exec(context.Background(), query)

	if err != nil {
		fmt.Fprintf(os.Stderr, "aaksdjf %v\n", err)
	}

	if commandTag.RowsAffected() != 1 {
		err := errors.New("Film is not updated.")
		return err
	}
	if len(actorList) > 0 {
		for _, actor_id := range actorList {
			commandTag, err := r.db.Exec(context.Background(), "INSERT INTO public.actor_film(actor_id, film_id) VALUES ($1, $2);", actor_id, filmId)
			if err != nil {
				return err
				log.Fatal(err)
			}
			if commandTag.RowsAffected() != 1 {
				log.Fatal(errors.New("Film->Actor not created"))
			}
		}
	}
	return err
}
func initDB(db *pgx.Conn) {
	query, err := ioutil.ReadFile("migrate.sql")
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}
	db.Exec(context.Background(), string(query))
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
	initDB(conn)
	postgresRepo := NewPostgresRepository(conn)
	handler := NewHandlerController(postgresRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /actor", handler.getAllActors)
	mux.HandleFunc("POST /actor", handler.createActor)
	mux.HandleFunc("PUT /actor", handler.updateActor)
	mux.HandleFunc("DELETE /actor/{id}", handler.deleteActor)
	mux.HandleFunc("POST /film", handler.createFilm)
	mux.HandleFunc("PUT /film", handler.updateFilm)
	mux.HandleFunc("DELETE /film/{id}", handler.deleteFilm)
	mux.HandleFunc("GET /films_sort_by/{field_name}/{order}", handler.getSortedFilms)
	mux.HandleFunc("GET /films", handler.getFilms)
	mux.HandleFunc("GET /search_film/by_film_name/{part_of_name}", handler.getFilmsByName)
	//mux.HandleFunc("GET /search_film/by_actor_name/{part_of_name}", handler.getFilmsByActorName)

	err = http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
	}
}
