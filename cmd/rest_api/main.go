package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/osiaeg/go_rest_api/internal/config"
	"github.com/osiaeg/go_rest_api/internal/database/postgresql"
	"github.com/osiaeg/go_rest_api/internal/transport/rest"
)

func main() {
	var env string
	flag.StringVar(&env, "env", "local", "The environment where the application will be launched.")
	flag.Parse()

	cfg := config.Parse(env)

	conn, err := pgx.Connect(context.Background(), cfg.Database.Url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	postgresql.InitDB(conn)
	postgresRepo := postgresql.NewPostgresRepository(conn)
	handler := rest.NewHandlerController(postgresRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /actor", handler.GetAllActors)
	mux.HandleFunc("POST /actor", handler.CreateActor)
	mux.HandleFunc("PUT /actor", handler.UpdateActor)
	mux.HandleFunc("DELETE /actor/{id}", handler.DeleteActor)
	mux.HandleFunc("POST /film", handler.CreateFilm)
	mux.HandleFunc("PUT /film", handler.UpdateFilm)
	mux.HandleFunc("DELETE /film/{id}", handler.DeleteFilm)
	mux.HandleFunc("GET /films_sort_by/{field_name}/{order}", handler.GetSortedFilms)
	mux.HandleFunc("GET /films", handler.GetFilms)
	mux.HandleFunc("GET /search_film/by_film_name/{part_of_name}", handler.GetFilmsByName)
	//mux.HandleFunc("GET /search_film/by_actor_name/{part_of_name}", handler.getFilmsByActorName)

	fmt.Printf("Server launched.\nURL: http://%s:%s\n", cfg.Server.Host, cfg.Server.Port)

	err = http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port), mux)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
	}
}
