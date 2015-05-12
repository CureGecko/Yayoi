/*
Users.go
Yayoi

Created by Cure Gecko on 5/10/15.
Copyright 2015, Cure Gecko. All rights reserved.

User account management and login.
*/

package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

type Users struct {
	Server  Server
	Writer  http.ResponseWriter
	Request *http.Request
	Path    []string
}

func (u Users) Process() {
	switch u.Path[1] {
	case "login":
		u.Login()
	case "signup":
		u.Signup()
	case "forgot":
		u.Forgot()
	case "reset":
		u.Reset()
	default:
		fmt.Fprint(u.Writer, "Hello World\n")
	}
}

func (u Users) isHexadecimal(s string) bool {
	return regexp.MustCompile("^[0-9a-fA-F]+$").MatchString(s)
}

func (u Users) isEmail(s string) bool {
	return regexp.MustCompile("^([\\w-]+(?:\\.[\\w-]+)*)@((?:[\\w-]+\\.)*\\w[\\w-]{0,66})\\.([a-z]{2,6}(?:\\.[a-z]{2})?)$").MatchString(s)
}

func (u Users) Login() {
	/*
		Import
		"crypto/sha256"
		"golang.org/x/crypto/pbkdf2"

		result := pbkdf2.Key([]byte("Test"), []byte("Test"), 1000, 64, sha256.New)
		fmt.Printf("%x", result)
	*/
	t, _ := template.ParseFiles("resources/Users/login.html")
	t.Execute(u.Writer, nil)
}

func (u Users) Signup() {
	if u.Request.Form["email"] != nil {
		type Info struct {
			Success bool
			Message string
		}
		info := Info{true, ""}

		if !u.isEmail(u.Request.Form["email"][0]) {
			info.Success = false
			info.Message = "Invalid email address provided."
		}
		if u.Request.Form["passwordSalt"] == nil || !u.isHexadecimal(u.Request.Form["passwordSalt"][0]) || len(u.Request.Form["passwordSalt"][0]) != 32 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid salt provided."
		}
		if u.Request.Form["password"] == nil || !u.isHexadecimal(u.Request.Form["password"][0]) || len(u.Request.Form["password"][0]) != 128 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid password provided."
		}
		if u.Request.Form["name"] == nil || len(u.Request.Form["name"][0]) == 0 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid name provided."
		}
		if info.Success {
			var id int
			err := u.Server.db.QueryRow("SELECT id FROM users WHERE email=?", u.Request.Form["email"][0]).Scan(&id)
			if err != nil && err != sql.ErrNoRows {
				log.Fatal(err)
				info.Success = false
				info.Message = "There was an unexpected error."
			} else if err != sql.ErrNoRows {
				info.Success = false
				info.Message = "The email address is already in use."
			}
		}
		if info.Success {
			var id int
			err := u.Server.db.QueryRow("SELECT id FROM users WHERE name=?", u.Request.Form["name"][0]).Scan(&id)
			if err != nil && err != sql.ErrNoRows {
				log.Fatal(err)
				info.Success = false
				info.Message = "There was an unexpected error."
			} else if err != sql.ErrNoRows {
				info.Success = false
				info.Message = "The username is already in use."
			}
		}

		if info.Success {
			info.Message = "Everything looks ok!"
		}

		t, _ := template.ParseFiles("resources/Users/signup_result.html")
		t.Execute(u.Writer, info)
	} else {
		type Info struct {
			PasswordSalt string
		}
		b := make([]byte, 16)
		rand.Read(b)
		salt := fmt.Sprintf("%x", b)
		info := Info{salt}

		t, _ := template.ParseFiles("resources/Users/signup.html")
		t.Execute(u.Writer, info)
	}
}

func (u Users) Forgot() {
	t, _ := template.ParseFiles("resources/Users/forgot.html")
	t.Execute(u.Writer, nil)
}

func (u Users) Reset() {
	t, _ := template.ParseFiles("resources/Users/reset.html")
	t.Execute(u.Writer, nil)
}
