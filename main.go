package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const ( // server
	port string = "8080"
)

const ( // website
	minimumUsernameLength int    = 3
	maximumUsernameLength int    = 20
	usernameRegexp        string = "^[a-zA-Z0-9_]+$"

	minimumPasswordLength int = 8
	maximumPasswordLength int = 50

	recaptchaSecret string = "6LcNHVUnAAAAAICv4oKEzhh6UTHk3QraFDfdde01"
)

var ( // configurations
	whitelistOn bool     = false
	whitelist   []string = []string{}
)

const ( // debugging
	debugPrefix   string = "\033[0;33m[DEBUG]\033[0m"
	errorPrefix   string = "\033[0;91m[ERROR]\033[0m"
	warningPrefix string = "\033[0;93m[WARNING]\033[0m"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.OpenFile("accesslogs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(errorPrefix, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		_, err = file.WriteString(fmt.Sprintf("%s - - [%s:%s] %s %s - %s\n", strings.Split(r.RemoteAddr, ":")[0], time.Now().Format("02/Jan/2006"), time.Now().Format("15:04:05"), r.Method, r.URL.Path, r.UserAgent()))
		if err != nil {
			fmt.Println(errorPrefix, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if whitelistOn {
			var isAuthorized bool = false
			for _, v := range whitelist {
				if strings.Split(r.RemoteAddr, ":")[0] == v {
					isAuthorized = true
				}
			}
			if !isAuthorized {
				http.Error(w, "", http.StatusForbidden)
				return
			}
		}

		db, err := sql.Open("sqlite3", "data.db")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer db.Close()

		URLPath := strings.TrimSuffix(r.URL.Path, "/")
		switch URLPath {
		case "":
			http.Redirect(w, r, "/home", http.StatusSeeOther)
		case "/register":
			if r.Method == http.MethodGet {
				tmpl, err := template.ParseFiles("pages/register.html")
				if err != nil {
					fmt.Println(errorPrefix, err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				tmpl.Execute(w, nil)
			} else if r.Method == http.MethodPost {
				tmpl, err := template.ParseFiles("pages/register.html")
				if err != nil {
					fmt.Println(errorPrefix, err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				response := struct {
					Message string `json:"message"`
				}{}

				err = r.ParseForm()
				if err != nil {
					fmt.Println(errorPrefix, err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}

				match, err := regexp.MatchString(usernameRegexp, r.FormValue("username"))
				if err != nil {
					fmt.Println(errorPrefix, err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}

				good := func() bool { // check username and password fields
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
					return true
				}()

				if good {
					_, token, err := createUser(r.FormValue("username"), r.FormValue("password"), db)
					if err != nil {
						fmt.Println(errorPrefix, err)
						http.Error(w, "", http.StatusInternalServerError)
						return
					}
					http.SetCookie(w, &http.Cookie{
						Name:  "token",
						Value: token,
					})
					http.Redirect(w, r, "/home", http.StatusSeeOther)
				}

				tmpl.Execute(w, response)
			}
		case "/login":
			if r.Method == http.MethodGet {
				tmpl, err := template.ParseFiles("pages/login.html")
				if err != nil {
					fmt.Println(errorPrefix, err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				tmpl.Execute(w, nil)
			} else if r.Method == http.MethodPost {
				tmpl, err := template.ParseFiles("pages/login.html")
				if err != nil {
					fmt.Println(errorPrefix, err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				response := struct {
					Message string `json:"Message"`
				}{}

				r.ParseForm()

				row := db.QueryRow("SELECT id FROM users WHERE username = ?", r.FormValue("username"))
				var id int
				err = row.Scan(&id)
				if err == sql.ErrNoRows {
					response.Message = "Username or password incorrect."
					tmpl.Execute(w, response)
					return
				}
				if err != nil {
					fmt.Println(errorPrefix, err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}

				tokenValue, err := newToken()
				if err != nil {
					fmt.Println(errorPrefix, err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				_, err = db.Exec("INSERT OR REPLACE INTO tokens (id, token) VALUES (?, ?)", id, tokenValue)
				if err != nil {
					fmt.Println(errorPrefix, err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:  "token",
					Value: tokenValue,
				})
				http.Redirect(w, r, "/home", http.StatusSeeOther)

				tmpl.Execute(w, response)
			}
		case "/home":
			tmpl, err := template.ParseFiles("pages/home.html")
			if err != nil {
				fmt.Println(errorPrefix, err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			response := struct {
				Username string `json:"username"`
			}{}

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

			row := db.QueryRow("SELECT id FROM tokens WHERE token = ?", tokenCookie.Value)
			var id int
			err = row.Scan(&id)
			if err == sql.ErrNoRows {
				removeTokenCookie(w)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if err != nil {
				fmt.Println(errorPrefix, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			row = db.QueryRow("SELECT username FROM users WHERE id = ?", id)
			var username string
			err = row.Scan(&username)
			if err == sql.ErrNoRows {
				removeTokenCookie(w)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if err != nil {
				fmt.Println(errorPrefix, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			response.Username = username

			tmpl.Execute(w, response)
		default:
			http.Error(w, "", http.StatusNotFound)
			return
		}
	})
	fmt.Println("Started server on :" + port)
	fmt.Println(http.ListenAndServe(":"+port, nil))
}

type RecaptchaResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

func verifyRecaptcha(response string) bool {
	data := url.Values{}
	data.Set("secret", recaptchaSecret)
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

	var recaptchaResp RecaptchaResponse
	err = json.Unmarshal(body, &recaptchaResp)
	if err != nil {
		return false
	}

	return recaptchaResp.Success
}

func createUser(username, password string, db *sql.DB) (id int, token string, err error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", err
	}

	_, err = db.Exec("INSERT INTO users (username, password) VALUES(?, ?)", username, hashedBytes)
	if err != nil {
		return 0, "", err
	}

	row := db.QueryRow("SELECT id FROM users WHERE username = ?", username)
	err = row.Scan(&id)
	if err != nil {
		return 0, "", err
	}

	userToken, err := newToken()
	if err != nil {
		return 0, "", err
	}
	_, err = db.Exec("INSERT INTO tokens (id, token) VALUES(?, ?)", id, userToken)
	if err != nil {
		return 0, "", err
	}

	return id, userToken, nil
}

func newToken() (string, error) {
	var length int = 50
	token := make([]byte, length)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	tokenString := base64.URLEncoding.EncodeToString(token)

	return tokenString, nil
}

func removeTokenCookie(w http.ResponseWriter) {
	expiredToken := &http.Cookie{
		Name:    "token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(1, 0),
		MaxAge:  -1,
	}
	http.SetCookie(w, expiredToken)
}
