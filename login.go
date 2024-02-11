package main

import (
	"database/sql"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		buf, err := pages.ReadFile("pages/login.html")
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}
		var body string = string(buf)

		tmpl, err := template.New("login").Parse(body)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		tmpl.Execute(w, nil)
	case http.MethodPost:
		buf, err := pages.ReadFile("pages/login.html")
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}
		var body string = string(buf)

		tmpl, err := template.New("login").Parse(body)
		if err != nil {
			logError(err)
			giveError(w, err)
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

		id, err := idFromUsername(r.FormValue("username"))
		if err == sql.ErrNoRows {
			response.Message = "Username or password incorrect."
			tmpl.Execute(w, response)
			return
		}
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		user, err := userFromUsername(r.FormValue("username"), permSystem)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.FormValue("password"))); err == bcrypt.ErrMismatchedHashAndPassword {
			response.Message = "Username or password incorrect."
			tmpl.Execute(w, response)
			return
		} else if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		tokenValue, err := newToken(id)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		if err := addUserToken(id, tokenValue); err != nil {
			logError(err)
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
