package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/osiaeg/go_rest_api/internal/database/postgresql"
	"github.com/osiaeg/go_rest_api/internal/models"
)

// type Response struct {
// 	message string `json:'message'`
// }

type Response map[string]interface{}

type HandlerController struct {
	repo *postgresql.PostgresRepository
}

func sendResponse(w http.ResponseWriter, code int, msg string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(Response{
		"message": msg,
	})
	if err != nil {
		return err
	}
	log.Println(msg)
	return nil
}

func NewHandlerController(repo *postgresql.PostgresRepository) *HandlerController {
	return &HandlerController{repo: repo}
}

func (h *HandlerController) CreateActor(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s request %s", r.Method, r.URL.Path))

	var a models.Actor
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error parse input data.")
		return
	}

	err = h.repo.CreateActor(&a)
	if err != nil {
		http.Error(w, "Error while creating actor in databse.", http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	if err := sendResponse(w, http.StatusCreated, "Actor is created."); err != nil {
		log.Println(err)
	}
}

func (h *HandlerController) CreateFilm(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s request %s", r.Method, r.URL.Path))
	var f models.Film
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error parse input data.")
		return
	}
	err = h.repo.CreateFilm(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := sendResponse(w, http.StatusCreated, "Film is created."); err != nil {
		log.Println(err)
	}
}

func (h *HandlerController) UpdateActor(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s request %s", r.Method, r.URL.Path))
	var a models.Actor
	err := json.NewDecoder(r.Body).Decode(&a)
	if err != nil {
		if err := sendResponse(w, http.StatusInternalServerError, err.Error()); err != nil {
			log.Println(err.Error())
			return
		}
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
		msg := fmt.Sprintf("Actor with id=%d not found.", a.Id)
		if err := sendResponse(w, http.StatusBadRequest, msg); err != nil {
			log.Println("Error update actor")
			return
		}
		return
	}

	msg := fmt.Sprintf("Actor with id=%d updated.", a.Id)
	if err := sendResponse(w, http.StatusOK, msg); err != nil {
		log.Println("Error update actor")
	}
}

func (h *HandlerController) UpdateFilm(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s request %s", r.Method, r.URL.Path))
	var f models.Film
	f.Rating = -1
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		log.Println("Error update in databse.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msg := fmt.Sprintf("Film with id=%d updated.", f.Id)
	if err := sendResponse(w, http.StatusOK, msg); err != nil {
		log.Println("Error update film")
	}
}

func (h *HandlerController) GetAllActors(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s request %s", r.Method, r.URL.Path))
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
	log.Println(fmt.Sprintf("%s request %s", r.Method, r.URL.Path))
	actor_id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		if err := sendResponse(w, http.StatusBadRequest, "id must be integer."); err != nil {
			log.Println(err)
		}
		return
	}
	err = h.repo.DeleteActor(strconv.Itoa(actor_id))
	if err != nil {
		ans := fmt.Sprintf("Actor with id=%d not found.", actor_id)
		if err := sendResponse(w, http.StatusNotFound, ans); err != nil {
			log.Println(err)
		}
		return
	}

	ans := fmt.Sprintf("Actor with id=%d was deleted.", actor_id)
	w.WriteHeader(http.StatusNoContent)
	log.Println(ans)
}

func (h *HandlerController) DeleteFilm(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s request %s", r.Method, r.URL.Path))
	film_id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		if err := sendResponse(w, http.StatusBadRequest, "id must be integer."); err != nil {
			log.Println(err)
		}
		return
	}
	err = h.repo.DeleteFilm(strconv.Itoa(film_id))
	if err != nil {
		ans := fmt.Sprintf("Film with id=%d not found.", film_id)
		if err := sendResponse(w, http.StatusNotFound, ans); err != nil {
			log.Println(err)
		}
		return
	}
	ans := fmt.Sprintf("Film with id=%d was deleted.", film_id)
	w.WriteHeader(http.StatusNoContent)
	log.Println(ans)
}

func (h *HandlerController) GetSortedFilms(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s request %s", r.Method, r.URL.Path))
	field_name := r.PathValue("field_name")
	order := r.PathValue("order")

	w.Header().Set("Content-Type", "application/json")
	films := h.repo.GetSortedFilms(field_name, order)
	json.NewEncoder(w).Encode(films)
}

func (h *HandlerController) GetFilms(w http.ResponseWriter, r *http.Request) {
	log.Println(fmt.Sprintf("%s request %s", r.Method, r.URL.Path))
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
