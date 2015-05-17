/*
Auth.go
Yayoi

Created by Cure Gecko on 5/15/15.
Copyright 2015, Cure Gecko. All rights reserved.

Manages user authentication.
*/
package main

import (
	"log"
	"net/http"
	"time"
)

type Auth struct {
	Server         *Iori
	Writer         http.ResponseWriter
	Request        *http.Request
	Authenticated  bool
	Authentication *Authentication
	User           *User
}

func (a *Auth) Validate() {
	a.Authenticated = false
	now := time.Now().Unix()

	var token string
	for _, cookie := range a.Request.Cookies() {
		if cookie.Name == "IoriAuth" {
			token = cookie.Value
		}
	}

	obj, err := a.Server.DBmap.Get(Authentication{}, token)
	if err != nil {
		log.Println(err)
		return
	}
	if obj == nil {
		return
	}
	authentication := obj.(*Authentication)

	obj, err = a.Server.DBmap.Get(User{}, authentication.UserID)
	if err != nil {
		log.Println(err)
		return
	}
	if obj == nil {
		return
	}
	user := obj.(*User)

	a.Authenticated = true
	a.Authentication = authentication
	a.User = user

	if authentication.Expires <= now+(60*10) /* 10 minutes */ {
		authentication.Token = randomString(30)
		authentication.Expires = now + (60 * 60 * 24) /* 1day */
		_, err := a.Server.DBmap.Update(authentication)
		if err == nil {
			log.Println(err)
			return
		}

		IoriAuth := new(http.Cookie)
		IoriAuth.Name = "IoriAuth"
		IoriAuth.Value = authentication.Token
		IoriAuth.Path = SitePath
		IoriAuth.Expires = time.Now().Add(time.Hour * 24 /* 1 day */)
		IoriAuth.HttpOnly = true
		http.SetCookie(a.Writer, IoriAuth)
	}
}
