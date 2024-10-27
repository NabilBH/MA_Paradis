package model

type Node struct {
	Id         int         `yaml:"id"`
	Address    string      `yaml:"address"`
	Neighbours []Neighbour `yaml:"neighbours"`
}

type Neighbour struct {
	Id         int    `yaml:"id"`
	Address    string `yaml:"address"`
	EdgeWeight int    `yaml:"edge_weight"`
}
