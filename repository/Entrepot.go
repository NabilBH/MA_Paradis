package repository

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"hes-so.ch/gnutella/model"
)

type Entrepot struct {
	Path string
}

func (repo *Entrepot) GetItems() ([]model.Movie, error) {

	data, err := os.ReadFile(repo.Path)
	var movies []model.Movie

	if err != nil {
		return movies, err
	}

	decoder := yaml.NewDecoder(strings.NewReader(string(data)))

	decoder.KnownFields(true)

	err = decoder.Decode(&movies)

	if err != nil {
		fmt.Println("Error decoding Yaml")
		return movies, err
	}
	return movies, nil
}

func (repo *Entrepot) FindMoviesByTitle(title string) ([]model.Movie, error) {

	movies, err := repo.GetItems()
	var searchResult []model.Movie

	if err != nil {
		return searchResult, err
	}

	for _, m := range movies {
		if strings.EqualFold(m.Title, title) {
			searchResult = append(searchResult, m)
		}
	}

	return searchResult, nil

}
