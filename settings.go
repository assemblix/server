package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
)

func settings(w http.ResponseWriter, r *http.Request) {
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

		buf, err := pages.ReadFile("pages/settings.html")
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}
		var body string = string(buf)

		tmpl, err := template.New("settings").Parse(body)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		tmpl.Execute(w, nil)
	case http.MethodPost:
		buf, err := pages.ReadFile("pages/settings.html")
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}
		var body string = string(buf)

		tmpl, err := template.New("settings").Parse(body)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

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

		response := struct {
			Message string `json:"Message"`
		}{}

		if err := r.ParseForm(); err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		user, err := userFromToken(tokenCookie.Value, permSystem)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		if r.FormValue("username") != "" {
			_, err := idFromUsername(r.FormValue("username"))
			if err == sql.ErrNoRows {
				user.Username = r.FormValue("username")
			} else {
				response.Message = "That username already exists."
				tmpl.Execute(w, response)
				return
			}
		}

		if r.FormValue("password") != "" {
			if len(r.FormValue("password")) < minimumPasswordLength || len(r.FormValue("password")) > maximumPasswordLength {
				response.Message = fmt.Sprintf("Password minimum length: %s, Maximum length: %s\n", fmt.Sprint(minimumPasswordLength), fmt.Sprint(maximumPasswordLength))
				tmpl.Execute(w, response)
				return
			}
			user.Password = r.FormValue("password")
		}

		updateFromId(user.Id, user)

		response.Message = "Updated."
		tmpl.Execute(w, response)
	}
}
