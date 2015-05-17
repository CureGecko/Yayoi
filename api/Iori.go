/*
Iori.go
Yayoi

Created by Cure Gecko on 5/10/15.
Copyright 2015, Cure Gecko. All rights reserved.

Main server and request processor for Yayoi's API.
*/

package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mostafah/mandrill"
	"gopkg.in/gorp.v1"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"strings"
)

//The path on the web server that is this API.
const APIPath string = "/api/"

//The path of the main site.
const SitePath string = "/"

//MySQL Database Details.
const dbUser string = "root"
const dbPassword string = "password"
const dbName string = "yayoi"

//Main server structure for dealing with requests via FCGI.
type Iori struct {
	DBmap *gorp.DbMap
}

/*
Process the main information needed from a request and pass the request onto the proper processing structure.
Every processing structure should accept Server, Writer, Request, Path, and User.
This function is called by the FCGI listener.
*/
func (s *Iori) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//Set Iori Header
	writer.Header().Set("Iori", "The backend of Yayoi.")

	//Form data is not parsed automatically, we need to parse it so we can determine parameters passed in the request.
	request.ParseForm()

	//Check Authentication
	auth := new(Auth)
	auth.Server = s
	auth.Writer = writer
	auth.Request = request
	auth.Validate()

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
		Users{Server: s, Auth: auth, Writer: writer, Request: request, Path: path}.Process()
	default:
		fmt.Fprint(writer, "Hello World\n")
	}
}

/*
The main function of the program.
Starts up the FCGI server on port specified and adds the request processer above as the handler.
*/
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//Connect to MySQL Database.
	dbmap := initDb()
	defer dbmap.Db.Close()

	//API Key for Mandrill
	mandrill.Key = "API-Key"
	manErr := mandrill.Ping()
	if manErr != nil {
		log.Fatal(manErr)
	}

	//Setup FCGI Server
	iori := new(Iori)
	iori.DBmap = dbmap
	tcp, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatal(err)
	}
	fcgi.Serve(tcp, iori)
}
