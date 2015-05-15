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
			t.Execute(u.Writer, u.Auth.Name)
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
			fmt.Fprint(u.Writer, "Hello World\n")
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

	_, err := u.Server.DB.Exec("DELETE FROM authentications WHERE `userid`=? AND `token`=?", u.Auth.ID, u.Auth.Token)
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
			ID           int64
		}
		info := Info{true, false, "", false, 0}

		name := u.Request.Form.Get("name")
		providedPassword := u.Request.Form.Get("password")

		now := time.Now().Unix()
		remoteAddr := u.Request.RemoteAddr[0 : len(u.Request.RemoteAddr)-2]

		var loginNonce []byte
		var loginAttempts int64
		var lastAttempt int64
		err := u.Server.DB.QueryRow("SELECT `loginNonce`,`loginAttempts`,`lastAttempt` FROM login WHERE `ip`=?", remoteAddr).Scan(&loginNonce, &loginAttempts, &lastAttempt)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			info.Success = false
			info.UnknownError = true
		} else if err == sql.ErrNoRows {
			_, err := u.Server.DB.Exec("INSERT INTO login (`ip`,`loginNonce`,`loginAttempts`,`lastAttempt`) VALUES (?,'',1,?)", remoteAddr, now)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			}
		} else {
			if lastAttempt < now-(60*30) /* 30 minutes */ {
				loginAttempts = 0
			}
			if loginAttempts >= 5 {
				info.Success = false
				info.Message = "Too many login attempts within 30."
			}
			loginAttempts++
			_, err := u.Server.DB.Exec("UPDATE login SET `loginNonce`='',`loginAttempts`=?,`lastAttempt`=? WHERE `ip`=?", loginAttempts, now, remoteAddr)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			}
		}

		var id int64
		var password []byte
		var passwordSalt []byte
		var signupKey string
		var apiKey string
		if info.Success {
			err := u.Server.DB.QueryRow("SELECT `id`,`password`,`passwordSalt`,`signupKey`,`apiKey` FROM users WHERE `name`=?", name).Scan(&id, &password, &passwordSalt, &signupKey, &apiKey)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			} else if err == sql.ErrNoRows {
				info.Success = false
				info.Message = "Invalid username/password."
			} else if len(signupKey) != 0 {
				info.Success = false
				info.Verification = true
				info.ID = id
			} else {
				if isHexadecimal(providedPassword) && len(providedPassword) == 128 {
					hash := fmt.Sprintf("%x", sha512.Sum512(append(password, loginNonce...)))

					if !strings.EqualFold(hash, providedPassword) {
						info.Success = false
						info.Message = "Invalid username/password."
					}
				} else {
					hash, err := scrypt.Key([]byte(providedPassword), passwordSalt, 16384, 8, 1, 64)
					if err != nil {
						fmt.Println(err)
						info.Success = false
						info.UnknownError = true
					} else if hex.EncodeToString(hash) != hex.EncodeToString(password) {
						info.Success = false
						info.Message = "Invalid username/password."
					}

				}
			}
		}

		if info.Success {
			token := randomString(30)

			_, err := u.Server.DB.Exec("INSERT INTO authentications (`userID`,`ip`,`token`,`time`,`expires`) VALUES (?,?,?,?,?)", id, remoteAddr, token, now, now+(60*60*24) /* 1 day */)
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

				u.Server.DB.Exec("UPDATE users SET `lastLoginTime`=? WHERE `id`=?", now, id)
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
	remoteAddr := u.Request.RemoteAddr[0 : len(u.Request.RemoteAddr)-2]

	var loginAttempts int
	err := u.Server.DB.QueryRow("SELECT `loginAttempts` FROM login WHERE `ip`=?", remoteAddr).Scan(&loginAttempts)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		info.Success = false
		info.UnknownError = true
	} else if err == sql.ErrNoRows {
		_, err := u.Server.DB.Exec("INSERT INTO login (`ip`,`loginNonce`,`loginAttempts`,`lastAttempt`) VALUES (?,?,0,?)", remoteAddr, b, now)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.UnknownError = true
		}
	} else {
		_, err := u.Server.DB.Exec("UPDATE login SET `loginNonce`=? WHERE `ip`=?", b, remoteAddr)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.UnknownError = true
		}
	}

	name := u.Request.Form.Get("name")

	var passwordSalt []byte
	err = u.Server.DB.QueryRow("SELECT `passwordSalt` FROM users WHERE `name`=?", name).Scan(&passwordSalt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		info.Success = false
		info.UnknownError = true
	} else if err == sql.ErrNoRows {
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
			var id int64
			err := u.Server.DB.QueryRow("SELECT `id` FROM users WHERE `email`=?", email).Scan(&id)
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
			var id int64
			err := u.Server.DB.QueryRow("SELECT `id` FROM users WHERE `name`=?", name).Scan(&id)
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
			passwordHex, _ := hex.DecodeString(password)
			passwordSaltHex, _ := hex.DecodeString(passwordSalt)
			result, err := u.Server.DB.Exec("INSERT INTO users (`name`,`email`,`password`,`passwordSalt`,`signupKey`,`joinTime`) VALUES (?,?,?,?,?,?)", name, email, passwordHex, passwordSaltHex, signupKey, now)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			} else {
				id, err := result.LastInsertId()
				if err != nil {
					log.Println(err)
					info.Success = false
					info.UnknownError = true
				} else {
					msg := mandrill.NewMessageTo(email, name)
					msg.FromEmail = "noreply@yayoi.se"
					msg.FromName = "Yayoi"
					msg.Subject = "Verify your email address."
					msg.Text = "Please verify your email address at http://127.0.0.1/users/verify?id=" + strconv.FormatInt(id, 10) + "&key=" + signupKey
					res, err := msg.Send(false)
					if err != nil || len(res) == 0 {
						log.Println("Mandrill Error:", err)
						info.Success = false
						info.Message = "There was an error sending an email."
						u.Server.DB.Exec("DELETE FROM users WHERE `id`=?", id)
						u.Server.DB.Exec("ALTER TABLE users AUTO_INCREMENT=?", id)
					} else if res[0].Status != "sent" {
						log.Println("Result", res[0].Status, res[0].RejectionReason, res[0].Id)
						info.Success = false
						info.Message = "There was an error sending an email."
						u.Server.DB.Exec("DELETE FROM users WHERE `id`=?", id)
						u.Server.DB.Exec("ALTER TABLE users AUTO_INCREMENT=?", id)
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
		var id int64
		err := u.Server.DB.QueryRow("SELECT `id` FROM users WHERE `email`=?", email).Scan(&id)
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
		var id int64
		err := u.Server.DB.QueryRow("SELECT `id` FROM users WHERE `name`=?", name).Scan(&id)
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
		var email string
		var name string
		var signupKey string
		err := u.Server.DB.QueryRow("SELECT `id`,`password`,`passwordSalt`,`signupKey`,`apiKey` FROM users WHERE `name`=?", name).Scan(&email, &name, &signupKey)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
			info.Success = false
			info.Reason = "There was an unexpected error."
		} else if err == sql.ErrNoRows {
			info.Success = false
			info.Reason = "There is no account for this id."
		} else if len(signupKey) == 0 {
			info.Success = false
			info.Reason = "This account is already verified."
		} else {
			msg := mandrill.NewMessageTo(email, name)
			msg.FromEmail = "noreply@yayoi.se"
			msg.FromName = "Yayoi"
			msg.Subject = "Verify your email address."
			msg.Text = "Please verify your email address at http://127.0.0.1/users/verify?id=" + userID + "&key=" + signupKey
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
				info.Email = email
				info.Name = name
			}
		}
	} else if userID == "" || signupKey == "" {
		info.Success = false
		info.Reason = "Invalid request."
	} else {
		var id int64
		var name string
		err := u.Server.DB.QueryRow("SELECT `id`,`name` FROM users WHERE `id`=? AND `signupKey`=?", userID, signupKey).Scan(&id, &name)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "No user found to verify. The user may already have been verified."
		} else {
			_, err := u.Server.DB.Exec("UPDATE users SET `signupKey`='' WHERE `id`=?", id)
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
	if u.Request.Form["email"] != nil {
		type Info struct {
			Success      bool
			UnknownError bool
			Message      string
		}
		info := Info{true, false, ""}

		email := u.Request.Form.Get("email")

		var id int64
		var name string
		err := u.Server.DB.QueryRow("SELECT `id`,`name` FROM users WHERE `email`=? AND `signupKey`=''", email).Scan(&id, &name)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Message = "There is no user registered with this email address."
		} else {
			resetKey := randomString(30)
			now := time.Now().Unix()

			_, err := u.Server.DB.Exec("UPDATE users SET `resetKey`=?, `resetRequestTime`=? WHERE id=?", resetKey, now, id)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.UnknownError = true
			} else {
				msg := mandrill.NewMessageTo(email, name)
				msg.FromEmail = "noreply@yayoi.se"
				msg.FromName = "Yayoi"
				msg.Subject = "Password reset."
				msg.Text = "To reet your password, click the following link http://127.0.0.1/users/reset?id=" + strconv.FormatInt(id, 10) + "&key=" + resetKey
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

		var id int64
		var name string
		if info.Success {
			err := u.Server.DB.QueryRow("SELECT `id`,`name` FROM users WHERE `id`=? AND `resetKey`=?", userID, resetKey).Scan(&id, &name)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Message = "No user found to reset or incorrect reset key."
			}
		}

		if info.Success {
			passwordHex, _ := hex.DecodeString(password)
			passwordSaltHex, _ := hex.DecodeString(passwordSalt)
			_, err := u.Server.DB.Exec("UPDATE users SET `resetKey`='',`resetRequestTime`=0,`password`=?,`passwordSalt`=? WHERE id=?", passwordHex, passwordSaltHex, id)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Message = "Unexpected error occured."
			} else {
				info.Message = "Successfully reset password for " + name + "."
			}
		}

		t, _ := template.ParseFiles("resources/Users/reset_result.html")
		t.Execute(u.Writer, info)
	} else {
		type Info struct {
			Success      bool
			Reason       string
			Name         string
			UserID       int64
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
			var id int64
			var name string
			err := u.Server.DB.QueryRow("SELECT `id`,`name` FROM users WHERE `id`=? AND `resetKey`=?", userID, resetKey).Scan(&id, &name)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Reason = "No user found to reset or incorrect reset key."
			} else {
				_, err := u.Server.DB.Exec("UPDATE users SET `signupKey`='' WHERE `id`=?", id)
				if err != nil {
					log.Println(err)
					info.Success = false
					info.Reason = "There was an unexpected error."
				} else {
					info.Name = name
					info.UserID = id
					info.ResetKey = resetKey
				}
			}
		}

		t, _ := template.ParseFiles("resources/Users/reset.html")
		t.Execute(u.Writer, info)
	}
}
