package main

import (
	"html/template"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
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
			logError(err)
			giveError(w, err)
			return
		}
		var body string = string(buf)

		tmpl, err := template.New("home").Parse(body)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		user, err := userFromToken(tokenCookie.Value, permUser)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		if err := tmpl.Execute(w, user); err != nil {
			logError(err)
			giveError(w, err)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
