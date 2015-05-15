/*
Auth.go
Yayoi

Created by Cure Gecko on 5/15/15.
Copyright 2015, Cure Gecko. All rights reserved.

Manages user authentication.
*/
package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Auth struct {
	Server        *Iori
	Writer        http.ResponseWriter
	Request       *http.Request
	Authenticated bool
	ID            int64
	Token         string
	Name          string
	Level         uint
}

func (a *Auth) Validate() {
	a.Authenticated = false
	now := time.Now().Unix()
	var token string
	var id int64
	for _, cookie := range a.Request.Cookies() {
		if cookie.Name == "IoriAuth" {
			token = cookie.Value
		} else if cookie.Name == "IoriAuthID" {
			id, _ = strconv.ParseInt(cookie.Value, 10, 64)
		}
	}

	var ip string
	var expires int64
	err := a.Server.DB.QueryRow("SELECT `ip`,`expires` FROM authentications WHERE `userID`=? AND `token`=? AND `expires`>=?", id, token, now).Scan(&ip, &expires)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Auth:", err)
	} else if err != sql.ErrNoRows {
		var name string
		var level uint
		err := a.Server.DB.QueryRow("SELECT `name`,`level` FROM users WHERE `id`=?", id).Scan(&name, &level)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Auth:", err)
		} else if err != sql.ErrNoRows {
			a.Authenticated = true
			a.ID = id
			a.Token = token
			a.Name = name
			a.Level = level

			if expires <= now+(60*10) /* 10 minutes */ {
				newToken := randomString(30)
				_, err := a.Server.DB.Exec("UPDATE authentications SET `token`=?, `expires`=? WHERE `userid`=? AND `token`=?", newToken, now+(60*60*24) /* 1 day */, id, token)
				if err != nil {
					log.Println("Auth:", err)
				} else {
					a.Token = newToken

					IoriAuth := new(http.Cookie)
					IoriAuth.Name = "IoriAuth"
					IoriAuth.Value = newToken
					IoriAuth.Path = SitePath
					IoriAuth.Expires = time.Now().Add(time.Hour * 24 /* 1 day */)
					IoriAuth.HttpOnly = true
					http.SetCookie(a.Writer, IoriAuth)

					IoriAuthID := new(http.Cookie)
					IoriAuthID.Name = "IoriAuthID"
					IoriAuthID.Value = strconv.FormatInt(id, 10)
					IoriAuthID.Path = SitePath
					IoriAuthID.Expires = time.Now().Add(time.Hour * 24 /* 1 day */)
					IoriAuthID.HttpOnly = true
					http.SetCookie(a.Writer, IoriAuthID)
				}
			}
		}
	}
}
