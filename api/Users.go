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

type Users struct {
	Server  *Iori
	Auth    *Auth
	Writer  http.ResponseWriter
	Request *http.Request
	Path    []string
}

func (u Users) Process() {
	if len(u.Path) == 1 {
		t, _ := template.ParseFiles("resources/Users/index.html")
		t.Execute(u.Writer, u.Auth)
	} else if u.Auth.Authenticated {
		switch u.Path[1] {
		case "logout":
			u.Logout()
		case "login":
			fallthrough
		case "signup":
			fallthrough
		case "verify":
			fallthrough
		case "forgot":
			fallthrough
		case "reset":
			t, _ := template.ParseFiles("resources/Users/signedIn.html")
			t.Execute(u.Writer, u.Auth.User.Name)
		default:
			http.NotFound(u.Writer, u.Request)
		}
	} else {
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
			fmt.Fprint(u.Writer, "404 Not Found\n")
		}
	}
}

func (u Users) Logout() {
	IoriAuth := new(http.Cookie)
	IoriAuth.Name = "IoriAuth"
	IoriAuth.Value = ""
	IoriAuth.Path = SitePath
	IoriAuth.Expires = time.Now()
	IoriAuth.HttpOnly = true
	http.SetCookie(u.Writer, IoriAuth)

	_, err := u.Server.DBmap.Delete(u.Auth.Authentication)
	if err != nil {
		log.Println(err)
	}

	t, _ := template.ParseFiles("resources/Users/logout.html")
	t.Execute(u.Writer, nil)
}

func (u Users) Login() {
	if u.Request.Form["name"] != nil {
		type Info struct {
			Success      bool
			UnknownError bool
			Message      string
			Verification bool
			ID           uint64
		}
		info := Info{true, false, "", false, 0}

		name := u.Request.Form.Get("name")
		providedPassword := u.Request.Form.Get("password")

		now := time.Now().Unix()
		remoteAddr := u.Request.RemoteAddr[0:strings.LastIndex(u.Request.RemoteAddr, ":")]

		var login *Login
		obj, err := u.Server.DBmap.Get(Login{}, remoteAddr)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.UnknownError = true
		} else if obj == nil {
			login = new(Login)
			login.Ip = remoteAddr
			login.LoginAttempts = 1
			login.LastAttempt = now
			err := u.Server.DBmap.Insert(login)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			}
		} else {
			login = obj.(*Login)
			if login.LastAttempt < now-(60*30) /* 30 minutes */ {
				login.LoginAttempts = 0
			}
			if login.LoginAttempts >= 5 {
				info.Success = false
				info.Message = "Too many login attempts within 30."
			}
			login.LoginAttempts++
			_, err := u.Server.DBmap.Update(login)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			}
		}

		var user User
		if info.Success {
			err = u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `name`=?", name)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			} else if err == sql.ErrNoRows {
				info.Success = false
				info.Message = "Invalid username/password."
			} else if len(user.SignupKey) != 0 {
				info.Success = false
				info.Verification = true
				info.ID = user.Id
			} else {
				if isHexadecimal(providedPassword) && len(providedPassword) == 128 {
					hash := fmt.Sprintf("%x", sha512.Sum512(append(user.Password, login.LoginNonce...)))

					if !strings.EqualFold(hash, providedPassword) {
						info.Success = false
						info.Message = "Invalid username/password."
					}
				} else {
					hash, err := scrypt.Key([]byte(providedPassword), user.PasswordSalt, 16384, 8, 1, 64)
					if err != nil {
						fmt.Println(err)
						info.Success = false
						info.UnknownError = true
					} else if hex.EncodeToString(hash) != hex.EncodeToString(user.Password) {
						info.Success = false
						info.Message = "Invalid username/password."
					}
				}
			}
		}

		if info.Success {
			token := randomString(30)

			authentication := new(Authentication)
			authentication.Token = token
			authentication.UserID = user.Id
			authentication.Ip = remoteAddr
			authentication.Time = now
			authentication.Expires = now + (60 * 60 * 24) /* 1 day */
			err = u.Server.DBmap.Insert(authentication)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			} else {
				IoriAuth := new(http.Cookie)
				IoriAuth.Name = "IoriAuth"
				IoriAuth.Value = token
				IoriAuth.Path = SitePath
				IoriAuth.Expires = time.Now().Add(time.Hour * 24 /* 1 day */)
				IoriAuth.HttpOnly = true
				http.SetCookie(u.Writer, IoriAuth)

				info.Message = "Successfully authenticated."

				user.LastLoginTime = now
				u.Server.DBmap.Update(&user)
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
		Success      bool
		UnknownError bool
		Salt         string
		Nonce        string
		Message      string
	}

	b := make([]byte, 32)
	rand.Read(b)
	nonce := hex.EncodeToString(b)
	info := Info{true, false, "", nonce, ""}

	now := time.Now().Unix()
	remoteAddr := u.Request.RemoteAddr[0:strings.LastIndex(u.Request.RemoteAddr, ":")]

	var login *Login
	obj, err := u.Server.DBmap.Get(Login{}, remoteAddr)
	if err != nil {
		log.Println(err)
		info.Success = false
		info.UnknownError = true
	} else if obj == nil {
		login = new(Login)
		login.Ip = remoteAddr
		login.LoginNonce = b
		login.LastAttempt = now
		err := u.Server.DBmap.Insert(login)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.UnknownError = true
		}
	} else {
		login = obj.(*Login)
		login.LoginNonce = b
		_, err := u.Server.DBmap.Update(login)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.UnknownError = true
		}
	}

	name := u.Request.Form.Get("name")

	var user User
	err = u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `name`=?", name)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		info.Success = false
		info.UnknownError = true
	} else if err == sql.ErrNoRows {
		info.Success = false
		info.Message = "Invalid username."
	} else {
		info.Salt = hex.EncodeToString(user.PasswordSalt)
	}

	t, _ := template.ParseFiles("resources/Users/salt.html")
	t.Execute(u.Writer, info)
}

func (u Users) Signup() {
	if u.Request.Form["email"] != nil {
		type Info struct {
			Success      bool
			UnknownError bool
			Message      string
		}
		info := Info{true, false, ""}
		email := u.Request.Form.Get("email")
		name := u.Request.Form.Get("name")
		passwordSalt := u.Request.Form.Get("passwordSalt")
		password := u.Request.Form.Get("password")

		if !isEmail(email) || len(email) > 100 {
			info.Success = false
			info.Message = "Invalid email address provided."
		}
		if passwordSalt == "" || !isHexadecimal(passwordSalt) || len(passwordSalt) != 64 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid salt provided."
		}
		if password == "" || !isHexadecimal(password) || len(password) != 128 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid password provided."
		}
		if len(name) < 2 || regexp.MustCompile("^[A-Za-z0-9]+[A-Za-z0-9_-]*$").MatchString(name) == false || len(name) > 50 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid name provided."
		}
		if info.Success {
			var user User
			err := u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `email`=?", email)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			} else if err != sql.ErrNoRows {
				info.Success = false
				info.Message = "The email address is already in use."
			}
		}
		if info.Success {
			var user User
			err := u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `name`=?", email)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			} else if err != sql.ErrNoRows {
				info.Success = false
				info.Message = "The username is already in use."
			}
		}

		if info.Success {
			signupKey := randomString(30)
			now := time.Now().Unix()
			passwordBin, _ := hex.DecodeString(password)
			passwordSaltBin, _ := hex.DecodeString(passwordSalt)

			user := new(User)
			user.Name = name
			user.Email = email
			user.Password = passwordBin
			user.PasswordSalt = passwordSaltBin
			user.SignupKey = signupKey
			user.JoinTime = now
			err := u.Server.DBmap.Insert(user)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			} else {
				msg := mandrill.NewMessageTo(email, name)
				msg.FromEmail = "noreply@yayoi.se"
				msg.FromName = "Yayoi"
				msg.Subject = "Verify your email address."
				msg.Text = "Please verify your email address at " + genURL(u.Request.URL, "users/verify?id="+strconv.FormatUint(user.Id, 10)+"&key="+signupKey)
				res, err := msg.Send(false)
				if err != nil || len(res) == 0 {
					log.Println("Mandrill Error:", err)
					info.Success = false
					info.Message = "There was an error sending an email."
					u.Server.DBmap.Delete(user)
					u.Server.DBmap.Exec("ALTER TABLE users AUTO_INCREMENT=?", user.Id)
				} else if res[0].Status != "sent" {
					log.Println("Result", res[0].Status, res[0].RejectionReason, res[0].Id)
					info.Success = false
					info.Message = "There was an error sending an email."
					u.Server.DBmap.Delete(user)
					u.Server.DBmap.Exec("ALTER TABLE users AUTO_INCREMENT=?", user.Id)
				} else {
					info.Message = "Sucessfully created account. Check your email for an activation link."
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
		var user User
		err := u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `email`=?", email)
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
		var user User
		err := u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `email`=?", email)
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
		Success            bool
		VerificationResent bool
		Email              string
		Name               string
		Reason             string
	}
	info := Info{true, false, "", "", ""}

	userID := u.Request.Form.Get("id")
	signupKey := u.Request.Form.Get("key")
	if u.Request.Form["resend"] != nil && userID != "" {
		obj, err := u.Server.DBmap.Get(User{}, userID)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			info.Success = false
			info.Reason = "There was an unexpected error."
		} else if obj == nil {
			info.Success = false
			info.Reason = "There is no account for this id."
		} else if len(obj.(*User).SignupKey) == 0 {
			info.Success = false
			info.Reason = "This account is already verified."
		} else {
			user := obj.(*User)
			msg := mandrill.NewMessageTo(user.Email, user.Name)
			msg.FromEmail = "noreply@yayoi.se"
			msg.FromName = "Yayoi"
			msg.Subject = "Verify your email address."
			msg.Text = "Please verify your email address at " + genURL(u.Request.URL, "users/verify?id="+strconv.FormatUint(user.Id, 10)+"&key="+user.SignupKey)
			res, err := msg.Send(false)
			if err != nil || len(res) == 0 {
				log.Println("Mandrill Error:", err)
				info.Success = false
				info.Reason = "There was an error sending an email."
			} else if res[0].Status != "sent" {
				log.Println("Result", res[0].Status, res[0].RejectionReason, res[0].Id)
				info.Success = false
				info.Reason = "There was an error sending an email."
			} else {
				//Yes. Printing the email address can be exploited, but people will usually validate very quickly.
				info.VerificationResent = true
				info.Email = user.Email
				info.Name = user.Name
			}
		}
	} else if userID == "" || signupKey == "" {
		info.Success = false
		info.Reason = "Invalid request."
	} else {
		var user User
		err := u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `id`=? AND `SignupKey`=?", userID, signupKey)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "No user found to verify. The user may already have been verified."
		} else {
			user.SignupKey = ""
			_, err = u.Server.DBmap.Update(&user)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Reason = "There was an unexpected error."
			} else {
				info.Name = user.Name
			}
		}
	}

	t, _ := template.ParseFiles("resources/Users/verify.html")
	t.Execute(u.Writer, info)
}

func (u Users) Forgot() {
	if u.Request.Form["email"] != nil {
		type Info struct {
			Success      bool
			UnknownError bool
			Message      string
		}
		info := Info{true, false, ""}

		email := u.Request.Form.Get("email")

		var user User
		err := u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `email`=? AND `signupKey`=''", email)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Message = "There is no user registered with this email address."
		} else {
			user.ResetKey = randomString(30)
			user.ResetRequestTime = time.Now().Unix()

			_, err = u.Server.DBmap.Update(&user)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			} else {
				msg := mandrill.NewMessageTo(user.Email, user.Name)
				msg.FromEmail = "noreply@yayoi.se"
				msg.FromName = "Yayoi"
				msg.Subject = "Password reset."
				msg.Text = "To reet your password, click the following link " + genURL(u.Request.URL, "users/reset?id="+strconv.FormatUint(user.Id, 10)+"&key="+user.ResetKey)
				res, err := msg.Send(false)
				if err != nil || len(res) == 0 {
					log.Println("Mandrill Error:", err)
					info.Success = false
					info.Message = "There was an error sending an email."
				} else if res[0].Status != "sent" {
					log.Println("Result", res[0].Status, res[0].RejectionReason, res[0].Id)
					info.Success = false
					info.Message = "There was an error sending an email."
				} else {
					info.Message = "Successfully sent reset email to " + email
				}
			}
		}

		t, _ := template.ParseFiles("resources/Users/forgot_result.html")
		t.Execute(u.Writer, info)
	} else {
		t, _ := template.ParseFiles("resources/Users/forgot.html")
		t.Execute(u.Writer, nil)
	}
}

func (u Users) Reset() {
	if u.Request.Form["password"] != nil {
		type Info struct {
			Success bool
			Message string
		}
		info := Info{true, ""}

		userID := u.Request.Form.Get("id")
		resetKey := u.Request.Form.Get("key")
		passwordSalt := u.Request.Form.Get("passwordSalt")
		password := u.Request.Form.Get("password")

		if userID == "" || resetKey == "" {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message = "Invalid request."
		}

		if passwordSalt == "" || !isHexadecimal(passwordSalt) || len(passwordSalt) != 64 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid salt provided."
		}
		if password == "" || !isHexadecimal(password) || len(password) != 128 {
			info.Success = false
			if info.Message != "" {
				info.Message += " "
			}
			info.Message += "Invalid password provided."
		}

		var user User
		if info.Success {
			err := u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `id`=? AND `resetKey`=?", userID, resetKey)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Message = "No user found to reset or incorrect reset key."
			}
		}

		if info.Success {
			passwordBin, _ := hex.DecodeString(password)
			passwordSaltBin, _ := hex.DecodeString(passwordSalt)

			user.Password = passwordBin
			user.PasswordSalt = passwordSaltBin
			user.ResetKey = ""
			user.ResetRequestTime = 0
			_, err := u.Server.DBmap.Update(&user)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Message = "Unexpected error occured."
			} else {
				info.Message = "Successfully reset password for " + user.Name + "."
			}
		}

		t, _ := template.ParseFiles("resources/Users/reset_result.html")
		t.Execute(u.Writer, info)
	} else {
		type Info struct {
			Success      bool
			Reason       string
			Name         string
			UserID       uint64
			ResetKey     string
			PasswordSalt string
		}
		b := make([]byte, 32)
		rand.Read(b)
		salt := hex.EncodeToString(b)
		info := Info{true, "", "", 0, "", salt}

		userID := u.Request.Form.Get("id")
		resetKey := u.Request.Form.Get("key")

		if userID == "" || resetKey == "" {
			info.Success = false
			info.Reason = "Invalid request."
		} else {
			var user User
			err := u.Server.DBmap.SelectOne(&user, "SELECT * FROM users WHERE `id`=? AND `resetKey`=?", userID, resetKey)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Reason = "No user found to reset or incorrect reset key."
			} else {
				info.Name = user.Name
				info.UserID = user.Id
				info.ResetKey = user.ResetKey
			}
		}

		t, _ := template.ParseFiles("resources/Users/reset.html")
		t.Execute(u.Writer, info)
	}
}
