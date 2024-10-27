package repository

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"hes-so.ch/gnutella/model"
)

type Node struct {
	Path string
}

func (n *Node) GetNodeConfig() (model.Node, error) {
	file, err := os.ReadFile(n.Path)
	var node model.Node

	if err != nil {
		return node, err
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(file)))

	decoder.KnownFields(true)

	err = decoder.Decode(&node)

	if err != nil {
		fmt.Println("Error decoding Yaml")
		return node, err
	}
	return node, nil
}
