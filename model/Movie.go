package model

type Movie struct {
	Id     int      `"yaml" : "id"`
	Title  string   `yaml: "title"`
	Genres []string `yaml: "genres"`
	Path   string   `"yaml": "path"`
}
