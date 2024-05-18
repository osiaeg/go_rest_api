package models

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
