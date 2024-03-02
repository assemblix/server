package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	permUser int = iota
	permAdmin
	permSystem
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

func createUser(username, password string, cash int, admin bool, db *sql.DB) (id int, token string, err error) {
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
	_, err = db.Exec("INSERT OR REPLACE INTO tokens (id, token) VALUES(?, ?)", id, userToken)
	if err != nil {
		return 0, "", err
	}

	_, err = db.Exec("INSERT OR REPLACE INTO userdata (id, cash, admin) VALUES (?, ?, ?)", id, cash, admin)
	if err != nil {
		return 0, "", err
	}

	return id, userToken, nil
}
func addUserToken(id int, token string) error {
	_, err := db.Exec("INSERT OR REPLACE INTO tokens (id, token) VALUES(?, ?)", id, token)
	return err
}

// type recaptchaResponse struct {
// 	Success    bool     `json:"success"`
// 	ErrorCodes []string `json:"error-codes"`
// }

// func verifyRecaptcha(response, secret string) bool {
// 	data := url.Values{}
// 	data.Set("secret", secret)
// 	data.Set("response", response)

// 	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", data)
// 	if err != nil {
// 		logError(err)
// 		return false
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logError(err)
// 		return false
// 	}

// 	var recaptchaResp recaptchaResponse
// 	err = json.Unmarshal(body, &recaptchaResp)
// 	if err != nil {
// 		logError(err)
// 		return false
// 	}

// 	return recaptchaResp.Success
// }

type apiUserObject struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Cash     int    `json:"cash"`
	Password string `json:"password"`
}
type apiErrorObject struct {
	Message string `json:"message"`
}

func userFromToken(token string, permission int) (apiUserObject, error) {
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

	if permission >= permSystem {
		password, err := passwordFromId(id)
		if err != nil {
			return apiUserObject{}, err
		}
		object.Password = password
	}

	return object, nil
}
func userFromId(id, permission int) (apiUserObject, error) {
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

	if permission >= permSystem {
		password, err := passwordFromId(id)
		if err != nil {
			return apiUserObject{}, err
		}
		object.Password = password
	}

	return object, nil
}
func userFromUsername(username string, permission int) (apiUserObject, error) {
	var object apiUserObject

	id, err := idFromUsername(username)
	if err != nil {
		return apiUserObject{}, err
	}

	object.Id = id

	cash, err := cashFromId(id)
	if err != nil {
		return apiUserObject{}, err
	}
	object.Cash = cash

	if permission >= permSystem {
		password, err := passwordFromId(id)
		if err != nil {
			return apiUserObject{}, err
		}
		object.Password = password
	}

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

func passwordFromId(id int) (string, error) {
	row := db.QueryRow("SELECT password FROM users WHERE id = ?", id)
	var password string
	err := row.Scan(&password)

	return password, err
}

func updateFromId(id int, user apiUserObject) error {
	og, err := userFromId(id, permSystem)
	if err != nil {
		return err
	}

	if user.Username != og.Username {
		_, err := db.Exec("UPDATE users SET username = ? WHERE id = ?", user.Username, id)
		if err != nil {
			return err
		}
	}
	if user.Cash != og.Cash {
		_, err := db.Exec("UPDATE userdata SET cash = ? WHERE id = ?", user.Cash, id)
		if err != nil {
			return err
		}
	}

	if user.Password != "" {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedBytes, id)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func updateFromToken(token string, user apiUserObject) error {
	og, err := userFromToken(token, permSystem)
	if err != nil {
		return err
	}

	id, err := idFromToken(token)
	if err != nil {
		return err
	}

	if user.Username != og.Username {
		_, err := db.Exec("UPDATE users SET username = ? WHERE id = ?", user.Username, id)
		if err != nil {
			return err
		}
	}
	if user.Cash != og.Cash {
		_, err := db.Exec("UPDATE userdata SET cash = ? WHERE id = ?", user.Cash, id)
		if err != nil {
			return err
		}
	}

	if user.Password != "" {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedBytes)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}
