package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
)

func register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		buf, err := pages.ReadFile("pages/register.html")
		if err != nil {
			logError(err)
			giveError(w, err)
		}
		var body string = string(buf)

		tmpl, err := template.New("register").Parse(body)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		if err := tmpl.Execute(w, nil); err != nil {
			logError(err)
			giveError(w, err)
			return
		}
	case http.MethodPost:
		response := struct {
			Message string `json:"message"`
		}{}

		buf, err := pages.ReadFile("pages/register.html")
		if err != nil {
			logError(err)
			giveError(w, err)
		}
		var body string = string(buf)

		tmpl, err := template.New("register").Parse(body)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		if err := r.ParseForm(); err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		match, err := regexp.MatchString(usernameRegexp, r.FormValue("username"))
		if err != nil {
			logError(err)
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
			_, err := idFromUsername(r.FormValue("username"))
			if err != sql.ErrNoRows {
				response.Message = "That username already exists."
				return false
			}

			// if !verifyRecaptcha(r.FormValue("g-recaptcha-response"), recaptchaSecret) {
			// 	response.Message = "reCAPTCHA verification failed."
			// 	return false
			// }

			return true
		}(); valid {
			_, token, err := createUser(r.FormValue("username"), r.FormValue("password"), joinCash, false, db)
			if err != nil {
				logError(err)
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
			logError(err)
			giveError(w, err)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
