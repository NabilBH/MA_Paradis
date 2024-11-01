package Services

import (
	"bufio"
	"io"
	"log"
	"net"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"hes-so.ch/gnutella/model"
	"hes-so.ch/gnutella/repository"
)

var PORT string = ":3002"

type Server struct {
	Logger             *log.Logger
	Name               string
	entrepotRepository repository.Entrepot
	nodeConfig         model.Node
}

func contains(slice []string, elem string) bool {
	for _, item := range slice {
		if item == elem {
			return true
		}
	}
	return false
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

func (server *Server) NextHop(query model.Query) (string, bool) {
	for i, node := range query.Path {
		if server.nodeConfig.Address == node {
			return query.Path[i-1], true
		}
	}
	return "", false
}

func (server *Server) InitiateQuery(movieTitle string, ttl int) {
	localAddress := server.nodeConfig.Address
	newId := uuid.New().String()

	query := model.Query{
		Id:            newId,
		TTL:           ttl,
		Data:          movieTitle,
		Type:          1,
		SourceAddress: localAddress,
		Path:          []string{localAddress}}
	server.Logger.Printf("INFO: %v Searching for movie %v", server.Name, movieTitle)

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
		//Handle concurrent connections
		go server.HandleConnection(conn)
	}
}

func (server *Server) HandleQuery(receivedQuery model.Query) {
	// If query is a request for a movie
	// Chekc that the request did not loopback to inital send
	if receivedQuery.Type == 1 {
		server.Logger.Printf(
			"INFO: %v received a request from :%v (TTL : %v)",
			server.Name,
			receivedQuery.SourceAddress,
			receivedQuery.TTL,
		)

		//Search movie in entrepot
		movies, err := server.entrepotRepository.FindMoviesByTitle(receivedQuery.Data)
		if err != nil {
			server.Logger.Fatalf("Fatal : %v Could not fetch movies, Error:%v", server.Name, err)
		}
		newHop := append(receivedQuery.Path, server.nodeConfig.Address)
		//Matching Movie found
		if len(movies) > 0 {
			server.Logger.Printf("INFO: %v has the requested movie", server.Name)

			response := model.Query{
				Id:            receivedQuery.Id,
				TTL:           -1,
				Data:          server.nodeConfig.Address, // Information about the node containing the movie
				Type:          0,
				SourceAddress: server.nodeConfig.Address,
				Path:          newHop,
			}
			//send a response to the orignal sender
			go server.SendQuery(response, receivedQuery.SourceAddress)
		}
		// Forward request to all neighbors when TTL is not expired
		if receivedQuery.TTL-1 > 0 {
			forwardQuery := model.Query{
				Id:            receivedQuery.Id,
				TTL:           receivedQuery.TTL - 1,
				Data:          receivedQuery.Data,
				SourceAddress: server.nodeConfig.Address,
				Type:          1,
				Path:          newHop,
			}
			for _, neighbour := range server.nodeConfig.Neighbours {
				//Exclude the original request source
				if contains(receivedQuery.Path, neighbour.Address) {
					continue
				}
				go server.SendQuery(forwardQuery, neighbour.Address)

			}
		} else {
			server.Logger.Printf("Debug: TTL expired for request id: %v at %v", receivedQuery.Id, server.Name)
		}
		//the received query is a response from another node
	} else if receivedQuery.Type == 0 {
		server.Logger.Printf("INFO: %v has received a response from %v", server.Name, receivedQuery.SourceAddress)
		// if the node is the iniator stop sending back the response
		if receivedQuery.Path[0] == server.nodeConfig.Address {
			server.Logger.Printf(
				"SUCCESS: %v received response: node %v has the requested content with path %v \n",
				server.Name,
				receivedQuery.Data,
				receivedQuery.Path)
		} else {
			//Trace back the net hop
			sender, ok := server.NextHop(receivedQuery)
			receivedQuery.SourceAddress = server.nodeConfig.Address
			if ok {
				go server.SendQuery(receivedQuery, sender)
			}
		}
	}
}

func (server *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()
	var builder strings.Builder

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// End of data
			break
		}
		if err != nil {
			server.Logger.Fatalf("Fatal: %v error while reading data: %v\n", server.Name, err)
			return
		}
		builder.WriteString(line)
	}
	message := builder.String()

	var receivedQuery model.Query
	err := yaml.Unmarshal([]byte(message), &receivedQuery)

	if err != nil {
		server.Logger.Fatalf("Fatal : %v Could not unmarshal received message Error:%v", server.Name, err)
	}
	server.HandleQuery(receivedQuery)
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
		nodeConfig:         nodeConfig,
		entrepotRepository: repository.Entrepot{Path: nodeEntrepotPath},
		Logger:             logger,
	}

}
