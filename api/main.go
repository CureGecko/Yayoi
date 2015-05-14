/*
main.go
Yayoi

Created by Cure Gecko on 5/10/15.
Copyright 2015, Cure Gecko. All rights reserved.

Main server and request processor for Yayoi's API.
*/

package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mostafah/mandrill"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"strings"
)

//The path on the web server that is this API.
var APIPath string = "/api/"

//MySQL Database Details.
var dbUser string = "root"
var dbPassword string = "password"
var dbName string = "yayoi"

//Main server structure for dealing with requests via FCGI
type Server struct {
	DB *sql.DB
}

/*
Process the main information needed from a request and pass the request onto the proper processing structure.
Every processing structure should accept Server, Writer, Request, Path, and User.
This function is called by the FCGI listener.
*/
func (s Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//Form data is not parsed automatically, we need to parse it so we can determine parameters passed in the request.
	request.ParseForm()

	//Parse the path of the request.
	fullPath := strings.Replace(request.URL.Path, APIPath, "", 1)
	if len(fullPath) != 0 && fullPath[len(fullPath)-1:] == "/" {
		fullPath = fullPath[0 : len(fullPath)-1]
	}
	path := strings.Split(fullPath, "/")
	if len(fullPath) != 0 {
		path = Filter(path)
	}
	log.Println(path)

	//Determine which processor should process the request.
	switch path[0] {
	case "users":
		Users{s, writer, request, path}.Process()
	default:
		fmt.Fprint(writer, "Hello World\n")
	}
}

/*
The main function of the program.
Starts up the FCGI server on port specified and adds the request processer above as the handler.
*/
func main() {
	//Connect to MySQL Database.
	db, err := sql.Open("mysql", dbUser+":"+dbPassword+"@/"+dbName)
	if err != nil {
		log.Fatal(err)
	}
	//Close when the program closes.
	defer db.Close()

	//API Key for Mandrill
	mandrill.Key = "API-Key"
	manErr := mandrill.Ping()
	if manErr != nil {
		log.Fatal(manErr)
	}

	//Setup FCGI Server
	server := Server{db}
	tcp, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatal(err)
	}
	fcgi.Serve(tcp, server)
}
