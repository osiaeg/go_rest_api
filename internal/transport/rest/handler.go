package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/osiaeg/go_rest_api/internal/database/postgresql"
	"github.com/osiaeg/go_rest_api/internal/models"
)

type HandlerController struct {
	repo *postgresql.PostgresRepository
}

func NewHandlerController(repo *postgresql.PostgresRepository) *HandlerController {
	return &HandlerController{repo: repo}
}

func (h *HandlerController) CreateActor(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST request.")
	var a models.Actor
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.repo.CreateActor(&a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func (h *HandlerController) CreateFilm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST request.")
	var f models.Film
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(f)
	err = h.repo.CreateFilm(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func (h *HandlerController) UpdateActor(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PUT request.")
	var a models.Actor
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
	err = h.repo.UpdateActor(a.Id, updates)
	if err != nil {
		fmt.Println("alksdjfalksjdf")
	}
}
func (h *HandlerController) UpdateFilm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PUT request.")
	var f models.Film
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
	err = h.repo.UpdateFilm(f.Id, updates, f.ActorList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HandlerController) GetAllActors(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET request.")
	w.Header().Set("Content-Type", "application/json")
	actors := h.repo.GetAllActors()
	var actorsWithFilm []models.ActorWithFilms
	for _, actor := range actors {
		var actorWithFilm models.ActorWithFilms
		actorWithFilm.Id = actor.Id
		actorWithFilm.Name = actor.Name
		actorWithFilm.Sex = actor.Sex
		actorWithFilm.Birthday = actor.Birthday
		actorWithFilm.Films = h.repo.GetFilmsByActorId(actor.Id)
		actorsWithFilm = append(actorsWithFilm, actorWithFilm)
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(actorsWithFilm)
}

func (h *HandlerController) DeleteActor(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DELETE request.")
	actor_id := r.PathValue("id")
	err := h.repo.DeleteActor(actor_id)
	if err != nil {
		log.Fatal(err)
	}
}
func (h *HandlerController) DeleteFilm(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DELETE request.")
	film_id := r.PathValue("id")
	err := h.repo.DeleteFilm(film_id)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *HandlerController) GetSortedFilms(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET request.")
	field_name := r.PathValue("field_name")
	order := r.PathValue("order")

	w.Header().Set("Content-Type", "application/json")
	films := h.repo.GetSortedFilms(field_name, order)
	json.NewEncoder(w).Encode(films)
}

func (h *HandlerController) GetFilms(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET request.")
	w.Header().Set("Content-Type", "application/json")
	films := h.repo.GetSortedFilms("film_rating", "desc")
	json.NewEncoder(w).Encode(films)
}
func (h *HandlerController) GetFilmsByName(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET request.")
	w.Header().Set("Content-Type", "application/json")
	part_of_name := r.PathValue("part_of_name")
	films := h.repo.SearchFilmByName(part_of_name)
	json.NewEncoder(w).Encode(films)
}
