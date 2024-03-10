package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func apiv1(w http.ResponseWriter, r *http.Request) {
	var apiWriteNotFound = func(w http.ResponseWriter) {
		message, err := json.Marshal(apiErrorObject{
			Message: "Page Not Found",
		})
		if err != nil {
			logError(err)
			giveError(w, err)
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
			logError(err)
			giveError(w, err)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(message)
	}
	var apiWrite = func(w http.ResponseWriter, v any) {
		message, err := json.Marshal(v)
		if err != nil {
			logError(err)
			giveError(w, err)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(message)
	}

	var url []string = strings.Split( // /api/v1/user/ -> [user] && /api/v1/something/etc -> [something, etc]
		strings.TrimRight(
			strings.TrimLeft(
				strings.TrimLeft(
					r.URL.Path,
					"/"),
				"/api/v1"),
			"/"),
		"/")
	w.Header().Set("Content-Type", "application/json")

	switch url[0] {
	case "user":
		query := r.URL.Query()
		id, err := strconv.Atoi(query.Get("id"))
		if err != nil {
			apiWriteError(w, err)
			return
		}

		obj, err := userFromId(id, permUser)
		if err != nil {
			logError(err)
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
