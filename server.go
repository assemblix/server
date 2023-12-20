package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func server(w http.ResponseWriter, r *http.Request) {
	var url []string = strings.Split(strings.TrimRight(strings.TrimLeft(r.URL.Path, "/"), "/"), "/")

	w.Header().Set("Content-Type", "text/html")

	switch url[0] {
	case "":
		switch r.Method {
		case http.MethodGet:
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	case "register":
		switch r.Method {
		case http.MethodGet:
			buf, err := pages.ReadFile("pages/register.html")
			if err != nil {
				giveError(w, err)
			}
			var body string = string(buf)

			tmpl, err := template.New("register").Parse(body)
			if err != nil {
				giveError(w, err)
				return
			}

			if err := tmpl.Execute(w, nil); err != nil {
				giveError(w, err)
				return
			}
		case http.MethodPost:
			response := struct {
				Message string `json:"message"`
			}{}

			buf, err := pages.ReadFile("pages/register.html")
			if err != nil {
				giveError(w, err)
			}
			var body string = string(buf)

			tmpl, err := template.New("register").Parse(body)
			if err != nil {
				giveError(w, err)
				return
			}

			if err := r.ParseForm(); err != nil {
				giveError(w, err)
				return
			}

			match, err := regexp.MatchString(usernameRegexp, r.FormValue("username"))
			if err != nil {
				giveError(w, err)
				return
			}

			if valid := func() bool {
				if len(r.FormValue("username")) < minimumUsernameLength || len(r.FormValue("username")) > maximumUsernameLength || !match {
					response.Message = fmt.Sprintf("Username minimum length: %s, Maximum length: %s\n, Regexp: %s", fmt.Sprint(minimumUsernameLength), fmt.Sprint(maximumUsernameLength), usernameRegexp)
					return false
				}
				if len(r.FormValue("password")) < minimumPasswordLength || len(r.FormValue("password")) > maximumPasswordLength {
					response.Message = fmt.Sprintf("Password minimum length: %s, Maximum length: %s\n", fmt.Sprint(minimumPasswordLength), fmt.Sprint(maximumPasswordLength))
					return false
				}
				if r.FormValue("password") != r.FormValue("confirmPassword") {
					response.Message = "Passwords do not match."
					return false
				}
				row := db.QueryRow("SELECT id FROM users WHERE username = ?", r.FormValue("username"))
				err := row.Scan()
				if err != sql.ErrNoRows {
					response.Message = "That username already exists."
					return false
				}

				if !verifyRecaptcha(r.FormValue("g-recaptcha-response"), recaptchaSecret) {
					response.Message = "reCAPTCHA verification failed."
				}

				return true
			}(); valid {
				_, token, err := createUser(r.FormValue("username"), r.FormValue("password"), joinCash, db)
				if err != nil {
					giveError(w, err)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:  "token",
					Value: token,
				})

				http.Redirect(w, r, "/home", http.StatusSeeOther)
				return
			}

			if err := tmpl.Execute(w, response); err != nil {
				giveError(w, err)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	case "home":
		switch r.Method {
		case http.MethodGet:
			cookies := r.Cookies()
			var tokenCookie *http.Cookie
			for _, cookie := range cookies {
				if cookie.Name == "token" {
					tokenCookie = cookie
					break
				}
			}
			if tokenCookie == nil {
				http.Redirect(w, r, "/register", http.StatusSeeOther)
				return
			}

			buf, err := pages.ReadFile("pages/home.html")
			if err != nil {
				giveError(w, err)
			}
			var body string = string(buf)

			tmpl, err := template.New("home").Parse(body)
			if err != nil {
				giveError(w, err)
				return
			}

			user, err := userFromToken(tokenCookie.Value)
			if err != nil {
				giveError(w, err)
				return
			}

			if err := tmpl.Execute(w, user); err != nil {
				giveError(w, err)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	case "login":
		switch r.Method {
		case http.MethodGet:
			buf, err := pages.ReadFile("pages/login.html")
			if err != nil {
				giveError(w, err)
				return
			}
			var body string = string(buf)

			tmpl, err := template.New("login").Parse(body)
			if err != nil {
				fmt.Println(errorPrefix, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.Execute(w, nil)
		case http.MethodPost:
			buf, err := pages.ReadFile("pages/login.html")
			if err != nil {
				giveError(w, err)
				return
			}
			var body string = string(buf)

			tmpl, err := template.New("login").Parse(body)
			if err != nil {
				fmt.Println(errorPrefix, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			response := struct {
				Message string `json:"Message"`
			}{}

			if err := r.ParseForm(); err != nil {
				giveError(w, err)
				return
			}

			id, err := idFromUsername(r.FormValue("username"))
			if err == sql.ErrNoRows {
				response.Message = "Username or password incorrect."
				tmpl.Execute(w, response)
				return
			}
			if err != nil {
				giveError(w, err)
				return
			}

			tokenValue, err := newToken(id)
			if err != nil {
				giveError(w, err)
				return
			}

			if err := addUserToken(id, tokenValue); err != nil {
				giveError(w, err)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:  "token",
				Value: tokenValue,
			})

			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

func apiv1(w http.ResponseWriter, r *http.Request) {
	var apiWriteNotFound = func(w http.ResponseWriter) {
		message, err := json.Marshal(apiErrorObject{
			Message: "Page Not Found",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write(message)
	}
	var apiWriteError = func(w http.ResponseWriter, err error) {
		message, err := json.Marshal(apiErrorObject{
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(message)
	}
	var apiWrite = func(w http.ResponseWriter, v any) {
		message, err := json.Marshal(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(message)
	}

	var url []string = strings.Split(strings.TrimRight(strings.TrimLeft(strings.TrimLeft(r.URL.Path, "/"), "/api/v1"), "/"), "/")
	w.Header().Set("Content-Type", "application/json")

	switch url[0] {
	case "user":
		query := r.URL.Query()
		id, err := strconv.Atoi(query.Get("id"))
		if err != nil {
			apiWriteError(w, err)
			return
		}

		obj, err := userFromId(id)
		if err != nil {
			apiWriteError(w, err)
			return
		}

		apiWrite(w, obj)
	default:
		apiWriteNotFound(w)
		return
	}
}

func removeTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(1, 0),
		MaxAge:  -1,
	})
}

func giveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
