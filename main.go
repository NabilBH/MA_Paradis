package main

import (
	"log"
	"os"
	"time"

	"hes-so.ch/gnutella/Services"
)

func main() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	defer file.Close()

	// Initialize a logger instance
	logger := log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)

	//Used to handle parralele connections by the server
	// Init the each node
	server1 := Services.NewServer("node-1", logger)
	server2 := Services.NewServer("node-2", logger)
	server3 := Services.NewServer("node-3", logger)
	server4 := Services.NewServer("node-4", logger)
	server5 := Services.NewServer("node-5", logger)

	// each server will start listening
	go server1.Start()
	go server2.Start()
	go server3.Start()
	go server4.Start()
	time.Sleep(2 * time.Second) //Waiting all node to be ready

	server5.InitiateQuery("Batman", 5)

	time.Sleep(2 * time.Second)

}
