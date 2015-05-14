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
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/mostafah/mandrill"
	"golang.org/x/crypto/scrypt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//
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
	case "salt":
		u.Salt()
	case "signup":
		u.Signup()
	case "available":
		u.Available()
	case "verify":
		u.Verify()
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
	if u.Request.Form["name"] != nil {
		type Info struct {
			Success bool
			Message string
		}
		info := Info{true, ""}

		name := u.Request.Form.Get("name")
		providedPassword := u.Request.Form.Get("password")

		now := time.Now().Unix()
		remoteAddr := u.Request.RemoteAddr[0 : len(u.Request.RemoteAddr)-2]

		var loginNonce []byte
		var loginAttempts int64
		var lastAttempt int64
		err := u.Server.DB.QueryRow("SELECT loginNonce,loginAttempts,lastAttempt FROM login WHERE ip=?", remoteAddr).Scan(&loginNonce, &loginAttempts, &lastAttempt)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			info.Success = false
			info.Message = "There was an unexpected error."
		} else if err == sql.ErrNoRows {
			_, err = u.Server.DB.Exec("INSERT INTO login (`ip`,`loginNonce`,`loginAttempts`,`lastAttempt`) VALUES (?,'',1,?)", remoteAddr, now)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Message = "There was an unexpected error."
			}
		} else {
			if lastAttempt < now-(60*30) {
				loginAttempts = 0
			}
			if loginAttempts >= 5 {
				info.Success = false
				info.Message = "Too many login attempts within 30."
			}
			loginAttempts++
			_, err = u.Server.DB.Exec("UPDATE login SET loginNonce='',loginAttempts=?,lastAttempt=? WHERE ip=?", loginAttempts, now, remoteAddr)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Message = "There was an unexpected error."
			}
		}

		if info.Success {
			var id int
			var password []byte
			var passwordSalt []byte
			var apiKey string
			err := u.Server.DB.QueryRow("SELECT id,password,passwordSalt,apiKey FROM users WHERE name=?", name).Scan(&id, &password, &passwordSalt, &apiKey)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				info.Success = false
				info.Message = "There was an unexpected error."
			} else if err == sql.ErrNoRows {
				info.Success = false
				info.Message = "Invalid username/password."
			} else {
				if u.isHexadecimal(providedPassword) && len(providedPassword) == 128 {
					hash := fmt.Sprintf("%x", sha512.Sum512(append(password, loginNonce...)))

					if strings.EqualFold(hash, providedPassword) {
						info.Message = "Successfully authenticated."
					} else {
						info.Success = false
						info.Message = "Invalid username/password."
					}
				} else {
					hash, err := scrypt.Key([]byte(providedPassword), passwordSalt, 16384, 8, 1, 64)
					if err != nil {
						fmt.Println(err)
						info.Success = false
						info.Message = "There was an unexpected error."
					} else if hex.EncodeToString(hash) == hex.EncodeToString(password) {
						info.Message = "Successfully authenticated."
					} else {
						info.Success = false
						info.Message = "Invalid username/password."
					}

				}
			}
		}

		t, _ := template.ParseFiles("resources/Users/login_result.html")
		t.Execute(u.Writer, info)
	} else {
		t, _ := template.ParseFiles("resources/Users/login.html")
		t.Execute(u.Writer, nil)
	}
}

func (u Users) Salt() {
	type Info struct {
		Success bool
		Salt    string
		Nonce   string
		Message string
	}

	b := make([]byte, 32)
	rand.Read(b)
	nonce := hex.EncodeToString(b)
	info := Info{true, "", nonce, ""}

	now := time.Now().Unix()
	remoteAddr := u.Request.RemoteAddr[0 : len(u.Request.RemoteAddr)-2]

	var loginAttempts int
	err := u.Server.DB.QueryRow("SELECT loginAttempts FROM login WHERE ip=?", remoteAddr).Scan(&loginAttempts)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		info.Success = false
		info.Message = "There was an unexpected error."
	} else if err == sql.ErrNoRows {
		_, err = u.Server.DB.Exec("INSERT INTO login (`ip`,`loginNonce`,`loginAttempts`,`lastAttempt`) VALUES (?,?,0,?)", remoteAddr, b, now)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Message = "There was an unexpected error."
		}
	} else {
		_, err = u.Server.DB.Exec("UPDATE login SET loginNonce=? WHERE ip=?", b, remoteAddr)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Message = "There was an unexpected error."
		}
	}

	name := u.Request.Form.Get("name")

	var passwordSalt []byte
	userErr := u.Server.DB.QueryRow("SELECT passwordSalt FROM users WHERE name=?", name).Scan(&passwordSalt)
	if userErr != nil && userErr != sql.ErrNoRows {
		log.Println(userErr)
		info.Success = false
		info.Message = "There was an unexpected error."
	} else if userErr == sql.ErrNoRows {
		info.Success = false
		info.Message = "Invalid username."
	} else {
		info.Salt = hex.EncodeToString(passwordSalt)
	}

	t, _ := template.ParseFiles("resources/Users/salt.html")
	t.Execute(u.Writer, info)
}

func (u Users) Signup() {
	if u.Request.Form["email"] != nil {
		type Info struct {
			Success bool
			Message string
		}
		info := Info{true, ""}
		email := u.Request.Form.Get("email")
		name := u.Request.Form.Get("name")
		passwordSalt := u.Request.Form.Get("passwordSalt")
		password := u.Request.Form.Get("password")

		if !u.isEmail(email) || len(email) > 100 {
			info.Success = false
			info.Message = "Invalid email address provided."
		}
		if passwordSalt == "" || !u.isHexadecimal(passwordSalt) || len(passwordSalt) != 64 {
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
			err := u.Server.DB.QueryRow("SELECT id FROM users WHERE email=?", email).Scan(&id)
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
			err := u.Server.DB.QueryRow("SELECT id FROM users WHERE name=?", name).Scan(&id)
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
			result, err := u.Server.DB.Exec("INSERT INTO users (`name`,`email`,`password`,`passwordSalt`,`signupKey`,`joinTime`) VALUES (?,?,?,?,?,?)", name, email, passwordHex, passwordSaltHex, signupKey, now)
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
						u.Server.DB.Exec("DELETE FROM users WHERE id=?", id)
						u.Server.DB.Exec("ALTER TABLE users AUTO_INCREMENT=?", id)
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
		b := make([]byte, 32)
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
		err := u.Server.DB.QueryRow("SELECT id FROM users WHERE email=?", email).Scan(&id)
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
		err := u.Server.DB.QueryRow("SELECT id FROM users WHERE name=?", name).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			info.NameAvailable = false
			info.Reason = "There was an unexpected error."
		} else if err != sql.ErrNoRows {
			info.NameAvailable = false
			if info.Reason != "" {
				info.Reason += " "
			}
			info.Reason += "The username is already in use."
		}
	} else {
		info.NameAvailable = false
	}

	t, _ := template.ParseFiles("resources/Users/available.html")
	t.Execute(u.Writer, info)
}

func (u Users) Verify() {
	type Info struct {
		Success bool
		Name    string
		Reason  string
	}
	info := Info{true, "", ""}

	userID := u.Request.Form.Get("id")
	signupKey := u.Request.Form.Get("key")

	if userID == "" || signupKey == "" {
		info.Success = false
		info.Reason = "Invalid request."
	} else {
		var id int
		var name string
		err := u.Server.DB.QueryRow("SELECT id, name FROM users WHERE id=? AND signupKey=? COLLATE utf8_bin", userID, signupKey).Scan(&id, &name)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "No user found to verify. The user may already have been verified."
		} else {
			_, err := u.Server.DB.Exec("UPDATE users SET signupKey='' WHERE id=?", id)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Reason = "There was an unexpected error."
			} else {
				info.Name = name
			}
		}
	}

	t, _ := template.ParseFiles("resources/Users/verify.html")
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
