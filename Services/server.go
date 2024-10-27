package Services

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"hes-so.ch/gnutella/helpers"
	"hes-so.ch/gnutella/model"
	"hes-so.ch/gnutella/repository"
)

var PORT string = ":3002"

type Server struct {
	Logger             *log.Logger
	Name               string
	entrepotRepository repository.Entrepot
	history            map[string][]string
	nodeConfig         model.Node
	initiatedRequests  map[string]bool
}

func (server *Server) SendQuery(query model.Query, destAddress string) {

	outConn, err := net.Dial("tcp", destAddress+PORT)
	if err != nil {
		server.Logger.Fatalf("%v could not dial TCP, %v", server.Name, err)
		return
	}
	data, err := yaml.Marshal(&query)
	if err != nil {
		server.Logger.Fatalf("Fatal : %v Could not marshal query %v, Error:%v", server.Name, query, err)
	}
	if query.Type == 1 {
		server.Logger.Printf("INFO: %v sending query: %v to %v", server.Name, query, destAddress)

	} else if query.Type == 0 {
		server.Logger.Printf("INFO: %v sending response: %v to %v", server.Name, query, destAddress)
	}
	outConn.Write([]byte(data))
	defer outConn.Close()
}

func (server *Server) InitiateQuery(movieTitle string, ttl int) {
	localAddress := server.nodeConfig.Address

	query := model.Query{Id: "myUniqueId", TTL: ttl, Data: movieTitle, Type: 1, SourceAddress: localAddress}
	server.Logger.Printf("INFO: %v Searching for movie %v", server.Name, movieTitle)
	server.initiatedRequests[query.Id] = true
	for _, neighbour := range server.nodeConfig.Neighbours {
		go server.SendQuery(query, neighbour.Address)
	}
	//Wait for responses
	server.Start()
}

func (server *Server) Start() {

	localAddress := server.nodeConfig.Address

	//Start listening for messages
	ln, err := net.Listen("tcp", localAddress+PORT)

	if err != nil {
		server.Logger.Fatalf("Fatal: %v Failed to listen on TCP %v", server.Name, err)
		return
	}
	server.Logger.Printf("INFO: %v starts listening on %v", server.Name, localAddress)
	for {
		conn, _ := ln.Accept()
		defer conn.Close()
		//Read incoming
		var builder strings.Builder
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				// End of data
				break
			}
			if err != nil {
				fmt.Printf("Error reading data: %v\n", err)
				return
			}
			builder.WriteString(line)
		}
		message := builder.String()
		server.HandleMessage(message)
	}

}

func (server *Server) HandleMessage(message string) {

	var receivedQuery model.Query
	err := yaml.Unmarshal([]byte(message), &receivedQuery)

	if err != nil {
		server.Logger.Fatalf("Fatal : %v Could not unmarshal received message Error:%v", server.Name, err)
	}
	// If query is a request for a movie
	if receivedQuery.Type == 1 {
		server.Logger.Printf("INFO: %v received a request from :%v (TTL : %v)", server.Name, receivedQuery.SourceAddress, receivedQuery.TTL)

		if !helpers.InArray(receivedQuery.SourceAddress, server.history[receivedQuery.Id]) {
			server.history[receivedQuery.Id] = append(
				server.history[receivedQuery.Id],
				receivedQuery.SourceAddress)
		}

		//Search movie in entrepot
		movies, err := server.entrepotRepository.FindMoviesByTitle(receivedQuery.Data)
		if err != nil {
			server.Logger.Fatalf("Fatal : %v Could not fetch movies, Error:%v", server.Name, err)
		}

		//Matching Movie found
		if len(movies) > 0 {
			server.Logger.Printf("INFO: %v has the requested movie", server.Name)
			response := model.Query{
				Id:            receivedQuery.Id,
				TTL:           -1,
				Data:          server.nodeConfig.Address, // Information about the node containing the movie
				Type:          0,
				SourceAddress: server.nodeConfig.Address,
				Path:          server.Name,
			}
			//send a response to the orignal sender
			go server.SendQuery(response, receivedQuery.SourceAddress)
		}

		// decrement the ttl
		forwardQuery := model.Query{
			Id:            receivedQuery.Id,
			TTL:           receivedQuery.TTL - 1,
			Data:          receivedQuery.Data,
			SourceAddress: server.nodeConfig.Address,
			Type:          1,
		}
		// Forward request to all neighbors when TTL is not expired
		if forwardQuery.TTL > 0 {
			for _, neighbour := range server.nodeConfig.Neighbours {
				//Exclude the original request source
				if neighbour.Address != receivedQuery.SourceAddress {
					go server.SendQuery(forwardQuery, neighbour.Address)
				}
			}
		} else if receivedQuery.TTL == 0 {
			server.Logger.Printf("ERROR: TTL expired for request id: %v at %v", server.Name, receivedQuery.Id)
		}

		//the received query is a response from another node
	} else if receivedQuery.Type == 0 {
		server.Logger.Printf("INFO: %v has received a response from %v", server.Name, receivedQuery.SourceAddress)
		// if the node is the iniator stop sending bach the response
		if server.initiatedRequests[receivedQuery.Id] {
			server.Logger.Printf(
				"SUCCESS: %v received response: node %v has the requested content with path %v \n",
				server.Name,
				receivedQuery.Data,
				receivedQuery.Path)
		} else {
			receivedQuery.Path = server.Name + "-->" + receivedQuery.Path
			for _, sender := range server.history[receivedQuery.Id] {
				go server.SendQuery(receivedQuery, sender)
			}
			server.history[receivedQuery.Id] = []string{}
		}

	}
}

func NewServer(name string, logger *log.Logger) *Server {
	nodeEntrepotPath := filepath.Join("./nodes", name, "entrepot.yaml")
	nodeNeighborsPath := filepath.Join("./nodes", name, "/node.yaml")
	nodeRepository := repository.Node{Path: nodeNeighborsPath}
	nodeConfig, err := nodeRepository.GetNodeConfig()
	if err != nil {
		logger.Fatalf("FATAL : %v failed to load configuration%v", name, err)
	}

	return &Server{
		Name:               name,
		history:            make(map[string][]string),
		initiatedRequests:  make(map[string]bool),
		nodeConfig:         nodeConfig,
		entrepotRepository: repository.Entrepot{Path: nodeEntrepotPath},
		Logger:             logger,
	}

}
