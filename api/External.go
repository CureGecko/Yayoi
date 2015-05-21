/*
External.go
Yayoi

Created by Cure Gecko on 5/19/15.
Copyright 2015, Cure Gecko. All rights reserved.

Data pullers for external information.
*/
package main

import (
	"html/template"
	"log"
	"net/http"
)

type ExternalError struct {
	s string
}

func (e *ExternalError) Error() string {
	return e.s
}

func newExternalError(s string) *ExternalError {
	err := new(ExternalError)
	err.s = s
	return err
}

type ExternalData struct {
	Images     []string //The URL to the image.
	Post       string   //The URL to the post.
	AuthorName string
	AuthorURL  string
}

type ExternalProvider interface {
	Name() string          //Friendly provider name.
	Match(URL string) bool //Regex to match against the source URL.
	Get(URL string) (*ExternalData, error)
}

func externalParseURL(URL string, server *Iori) (*ExternalData, error) {
	providers := []ExternalProvider{
		&ExternalPixiv{Server: server},
		&ExternalNicoSeiga{Server: server},
		&ExternalBooru{Server: server},
		&ExternalSankakuComplex{Server: server},
	}

	for _, provider := range providers {
		if provider.Match(URL) {
			response, err := provider.Get(URL)
			return response, err
		}
	}
	return nil, newExternalError("No provider for url.")
}

type External struct {
	Server  *Iori
	Auth    *Auth
	Writer  http.ResponseWriter
	Request *http.Request
	Path    []string
}

func (e External) Process() {
	type Info struct {
		Success  bool
		Response *ExternalData
	}
	info := Info{}
	source := e.Request.Form.Get("source")

	response, err := externalParseURL(source, e.Server)
	if err != nil {
		log.Println(err)
		info.Success = false
	} else {
		info.Success = true
		info.Response = response
	}

	t, _ := template.ParseFiles("resources/external.html")
	t.Execute(e.Writer, info)
}
