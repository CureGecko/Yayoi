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
	"encoding/hex"
	"fmt"
	"github.com/mostafah/mandrill"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
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
	case "available":
		u.Available()
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
		email := u.Request.Form.Get("email")
		passwordSalt := u.Request.Form.Get("passwordSalt")
		password := u.Request.Form.Get("password")
		name := u.Request.Form.Get("name")

		if !u.isEmail(email) || len(email) > 100 {
			info.Success = false
			info.Message = "Invalid email address provided."
		}
		if passwordSalt == "" || !u.isHexadecimal(passwordSalt) || len(passwordSalt) != 32 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid salt provided."
		}
		if password == "" || !u.isHexadecimal(password) || len(password) != 128 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid password provided."
		}
		if name == "" || regexp.MustCompile("^[A-Za-z0-9]+[A-Za-z0-9_-]*$").MatchString(name) == false || len(name) > 50 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid name provided."
		}
		if info.Success {
			var id int
			err := u.Server.db.QueryRow("SELECT id FROM users WHERE email=?", email).Scan(&id)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				info.Success = false
				info.Message = "There was an unexpected error."
			} else if err != sql.ErrNoRows {
				info.Success = false
				info.Message = "The email address is already in use."
			}
		}
		if info.Success {
			var id int
			err := u.Server.db.QueryRow("SELECT id FROM users WHERE name=?", name).Scan(&id)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				info.Success = false
				info.Message = "There was an unexpected error."
			} else if err != sql.ErrNoRows {
				info.Success = false
				info.Message = "The username is already in use."
			}
		}

		if info.Success {
			signupKey := randomString(30)
			now := time.Now().Unix()
			passwordHex, _ := hex.DecodeString(password)
			passwordSaltHex, _ := hex.DecodeString(passwordSalt)
			result, err := u.Server.db.Exec("INSERT INTO users (`name`,`email`,`password`,`paswordSalt`,`signupKey`,`joinTime`) VALUES (?,?,?,?,?,?)", name, email, passwordHex, passwordSaltHex, signupKey, now)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Message = "There was an unexpected error."
			} else {
				id, err := result.LastInsertId()
				if err != nil {
					log.Println(err)
					info.Success = false
					info.Message = "There was an unexpected error."
				} else {
					msg := mandrill.NewMessageTo(email, name)
					msg.Text = "Please verify your email address at http://127.0.0.1/users/verify?id=" + strconv.FormatInt(id, 10) + "&key=" + signupKey
					msg.Subject = "Verify your email address."
					msg.FromEmail = "noreply@yayoi.se"
					msg.FromName = "Yayoi"
					res, err := msg.Send(false)
					if err != nil || len(res) == 0 {
						log.Println("Mandrill Error:", err)
						info.Success = false
						info.Message = "There was an error sending an email."
						u.Server.db.Exec("DELETE FROM users WHERE id=?", id)
						u.Server.db.Exec("ALTER TABLE users AUTO_INCREMENT=?", id)
					} else if res[0].Status != "sent" {
						log.Println("Result", res[0].Status, res[0].RejectionReason, res[0].Id)
						info.Success = false
						info.Message = "There was an error sending an email."
					} else {
						info.Message = "Sucessfully created account. Check your email for an activation link."
					}
				}
			}
		}

		t, _ := template.ParseFiles("resources/Users/signup_result.html")
		t.Execute(u.Writer, info)
	} else {
		type Info struct {
			PasswordSalt string
		}
		b := make([]byte, 16)
		rand.Read(b)
		salt := hex.EncodeToString(b)
		info := Info{salt}

		t, _ := template.ParseFiles("resources/Users/signup.html")
		t.Execute(u.Writer, info)
	}
}

func (u Users) Available() {
	type Info struct {
		EmailAvailable bool
		NameAvailable  bool
		Reason         string
	}
	info := Info{true, true, ""}

	name := u.Request.Form.Get("name")
	email := u.Request.Form.Get("email")

	if email != "" {
		var id int
		err := u.Server.db.QueryRow("SELECT id FROM users WHERE email=?", email).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			info.EmailAvailable = false
			info.Reason = "There was an unexpected error."
		} else if err != sql.ErrNoRows {
			info.EmailAvailable = false
			info.Reason = "The email address is already in use."
		}
	} else {
		info.EmailAvailable = false
	}

	if name != "" {
		var id int
		err := u.Server.db.QueryRow("SELECT id FROM users WHERE name=?", name).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			info.NameAvailable = false
			info.Reason = "There was an unexpected error."
		} else if err != sql.ErrNoRows {
			info.NameAvailable = false
			info.Reason = "The username is already in use."
		}
	} else {
		info.NameAvailable = false
	}

	t, _ := template.ParseFiles("resources/Users/available.html")
	t.Execute(u.Writer, info)
}

func (u Users) Forgot() {
	t, _ := template.ParseFiles("resources/Users/forgot.html")
	t.Execute(u.Writer, nil)
}

func (u Users) Reset() {
	t, _ := template.ParseFiles("resources/Users/reset.html")
	t.Execute(u.Writer, nil)
}
