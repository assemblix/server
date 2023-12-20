package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func newToken(id int) (string, error) {
	var length int = 50
	token := make([]byte, length)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	tokenString := "DO_NOT_SHARE_THIS" + "." + base64.URLEncoding.EncodeToString([]byte(fmt.Sprint(id))) + "." + base64.URLEncoding.EncodeToString(token)
	return tokenString, nil
}

func createUser(username, password string, cash int, db *sql.DB) (id int, token string, err error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", err
	}

	_, err = db.Exec("INSERT INTO users (username, password, joined) VALUES(?, ?, ?)", username, hashedBytes, time.Now().Unix())
	if err != nil {
		return 0, "", err
	}

	row := db.QueryRow("SELECT id FROM users WHERE username = ?", username)
	err = row.Scan(&id)
	if err != nil {
		return 0, "", err
	}

	userToken, err := newToken(id)
	if err != nil {
		return 0, "", err
	}
	_, err = db.Exec("INSERT INTO tokens (id, token) VALUES(?, ?)", id, userToken)
	if err != nil {
		return 0, "", err
	}

	_, err = db.Exec("INSERT INTO userdata (id, cash, admin) VALUES (?, ?, ?)", id, cash, false)
	if err != nil {
		return 0, "", err
	}

	return id, userToken, nil
}
func addUserToken(id int, token string) error {
	_, err := db.Exec("INSERT OR REPLACE INTO tokens (id, token) VALUES(?, ?)", id, token)
	return err
}

type recaptchaResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func verifyRecaptcha(response, secret string) bool {
	data := url.Values{}
	data.Set("secret", secret)
	data.Set("response", response)

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", data)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var recaptchaResp recaptchaResponse
	err = json.Unmarshal(body, &recaptchaResp)
	if err != nil {
		return false
	}

	return recaptchaResp.Success
}

type apiUserObject struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Cash     int    `json:"cash"`
}
type apiErrorObject struct {
	Message string `json:"message"`
}

func userFromToken(token string) (apiUserObject, error) {
	var object apiUserObject

	id, err := idFromToken(token)
	if err != nil {
		return apiUserObject{}, err
	}
	object.Id = id

	username, err := usernameFromId(id)
	if err != nil {
		return apiUserObject{}, err
	}
	object.Username = username

	cash, err := cashFromId(id)
	if err != nil {
		return apiUserObject{}, err
	}
	object.Cash = cash

	return object, nil
}
func userFromId(id int) (apiUserObject, error) {
	var object apiUserObject

	object.Id = id

	username, err := usernameFromId(id)
	if err != nil {
		return apiUserObject{}, err
	}
	object.Username = username

	cash, err := cashFromId(id)
	if err != nil {
		return apiUserObject{}, err
	}
	object.Cash = cash

	return object, nil
}

func idFromToken(token string) (int, error) {
	row := db.QueryRow("SELECT id FROM tokens WHERE token = ?", token)
	var id int
	err := row.Scan(&id)

	return id, err
}

func usernameFromId(id int) (string, error) {
	row := db.QueryRow("SELECT username FROM users WHERE id = ?", id)
	var username string
	err := row.Scan(&username)

	return username, err
}
func cashFromId(id int) (int, error) {
	row := db.QueryRow("SELECT cash FROM userdata WHERE id = ?", id)
	var cash int
	err := row.Scan(&cash)

	return cash, err
}

func idFromUsername(username string) (int, error) {
	row := db.QueryRow("SELECT id FROM users WHERE username = ?", username)
	var id int
	err := row.Scan(&id)

	return id, err
}
