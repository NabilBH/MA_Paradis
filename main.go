package main

import (
	"fmt"

	"hes-so.ch/gnutella/repository"
)

func main() {

	entrepotRepo := repository.Entrepot{Path: "./entrepots/entrepot.yaml"}
	//var movies []model.Movie
	movies, err := entrepotRepo.FindMoviesByTitle("Terminator")
	if err != nil {
		return
	}

	fmt.Printf("Movies Found : %+v\n", movies)

}
