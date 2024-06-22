package postgresql

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/osiaeg/go_rest_api/internal/models"
)

type PostgresRepository struct {
	db *pgx.Conn
}

func NewPostgresRepository(db *pgx.Conn) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateActor(m *models.Actor) error {
	commandTag, err := r.db.Exec(context.Background(), "insert into public.actor(actor_name, actor_sex, actor_birthday) values($1, $2, $3);", m.Name, m.Sex, m.Birthday)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "aaksdjf %v\n", err)
		return err
	}
	if commandTag.RowsAffected() != 1 {
		err := errors.New("Actor is not create.")
		return err
	}
	return err
}

func (r *PostgresRepository) SearchFilmByName(part_of_name string) []models.Film {
	query := fmt.Sprintf("select * from public.film where film_name like '%%%s%%'", part_of_name)
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var films []models.Film
	for rows.Next() {
		var f models.Film
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

func (r *PostgresRepository) DeleteActor(actor_id string) error {
	commandTag, err := r.db.Exec(context.Background(), "delete from public.actor where actor_id=$1", actor_id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aaksdjf %v\n", err)
	}
	if commandTag.RowsAffected() != 1 {
		err := errors.New("Actor is not deleted.")
		return err
	}
	//TODO: delete actor_id from film actor_list with this id
	return err
}

func (r *PostgresRepository) DeleteFilm(film_id string) error {
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

func (r *PostgresRepository) GetAllActors() []models.Actor {
	rows, err := r.db.Query(context.Background(), "select * from public.actor")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var actors []models.Actor
	for rows.Next() {
		var a models.Actor
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

func (r *PostgresRepository) GetSortedFilms(field_name string, order string) []models.Film {
	query := fmt.Sprintf("select * from public.film order by %s %s", field_name, order)
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var films []models.Film
	for rows.Next() {
		var f models.Film
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

func (r *PostgresRepository) GetFilmById(filmId int) models.Film {
	row := r.db.QueryRow(context.Background(), "select * from public.film where film_id=$1", filmId)

	var f models.Film
	var releaseDate time.Time
	err := row.Scan(&f.Id, &f.Name, &f.Description, &releaseDate, &f.Rating, &f.ActorList)
	f.ReleaseDate = fmt.Sprintf("%d-%02d-%02d", releaseDate.Year(), releaseDate.Month(), releaseDate.Day())
	if err != nil {
		log.Fatal(err)
	}
	return f
}

// TODO: use transaction, because if actor_id not found, film was created, but actor_film link not created.
func (r *PostgresRepository) CreateFilm(f *models.Film) error {
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

func prepareUpdateParameters(updates map[string]string) []string {
	var parameters []string

	for k, v := range updates {
		parameters = append(parameters, fmt.Sprintf("%s='%s'", k, v))
	}

	return parameters
}

func (r *PostgresRepository) UpdateActor(actorId int, updates map[string]string) error {
	parameters := prepareUpdateParameters(updates)

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

func (r *PostgresRepository) createActorToFilm(aId, fId int) error {
	query := "INSERT INTO public.actor_film(actor_id, film_id) VALUES ($1, $2);"
	commandTag, err := r.db.Exec(context.Background(), query, aId, fId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				log.Println("Get ForeignKeyViolation error")
				return errors.New(fmt.Sprintf("Actor with actor_id=%d not found.", aId))
			}
		}
		return err
	}

	if commandTag.RowsAffected() != 1 {
		log.Println("Film->Actor not created")
		return errors.New("Film->Actor not created")
	}
	return nil
}

func (r *PostgresRepository) deleteActorToFilm(aId, fId int) error {
	query := fmt.Sprintf("DELETE FROM public.actor_film WHERE actor_id=%d and film_id=%d", aId, fId)
	commandTag, err := r.db.Exec(context.Background(), query)
	if err != nil {
		log.Println("Error when exec delete.")
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return errors.New("Film->Actor not deleted.")
	}
	return nil
}

func (r *PostgresRepository) UpdateFilm(filmId int, updates map[string]string, actorList []int) error {
	var actrosIdOld []int
	actrosIdOld = r.GetActorIdByFilmId(filmId)
	var matchIndexes []int
	for _, actor_id := range actorList {

		if slices.Contains(actrosIdOld, actor_id) {
			matchIndexes = append(matchIndexes, slices.Index(actrosIdOld, actor_id))
			continue
		}

		err := r.createActorToFilm(actor_id, filmId)
		if err != nil {
			log.Println(err)
			return err
		}

	}
	for index, value := range actrosIdOld {
		if slices.Contains(matchIndexes, index) {
			continue
		}

		err := r.deleteActorToFilm(value, filmId)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	// Update row in public.film
	parameters := prepareUpdateParameters(updates)
	query := fmt.Sprintf("UPDATE public.film SET %s WHERE film_id=%d;", strings.Join(parameters, ", "), filmId)
	commandTag, err := r.db.Exec(context.Background(), query)
	if err != nil {
		log.Println("Error when exec")
	}

	if commandTag.RowsAffected() != 1 {
		err := errors.New("Film is not updated.")
		return err
	}
	// Update row in public.film
	return err
}

func (r *PostgresRepository) GetFilmsByActorId(actorId int) []models.Film {
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
	var films []models.Film
	for _, film_id := range filmsId {
		film := r.GetFilmById(film_id)
		films = append(films, film)
	}

	return films
}

func (r *PostgresRepository) GetActorIdByFilmId(filmId int) []int {
	rows, err := r.db.Query(context.Background(), "select actor_id from public.actor_film where film_id=$1", filmId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var actorsId []int
	for rows.Next() {
		var actorId int
		err := rows.Scan(&actorId)
		if err != nil {
			log.Fatal(err)
		}
		actorsId = append(actorsId, actorId)
	}

	return actorsId
}

func InitDB(db *pgx.Conn) {
	query, err := ioutil.ReadFile("../../migrations/migrate.sql")
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}
	db.Exec(context.Background(), string(query))
}
