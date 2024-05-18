package models

type Film struct {
	Id          int    `json:"film_id"`
	Name        string `json:"film_name"`
	Description string `json:"film_description"`
	ReleaseDate string `json:"film_release_date"`
	Rating      int    `json:"film_rating"`
	ActorList   []int  `json:"film_actor_list"` // Во время вывода передавать имена актеров, а не их id
}
