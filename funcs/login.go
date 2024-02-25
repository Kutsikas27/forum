package funcs

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/gofrs/uuid/v5"
	_ "github.com/mattn/go-sqlite3"
)

var sessions = []session{}

type session struct {
	UserUUID string
	UserName string
	expiry   time.Time
}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		HandleErrorPage(w, r)
		return
	}

	if r.Method == "POST" {
		database, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Println("Error opening database:", err)
			return
		}
		defer database.Close()
		switch r.FormValue("operation") {
		case "signup":
			if err := SignUp(w, database, r.FormValue("Name"), r.FormValue("Email"), r.FormValue("Password")); err != nil {
				if err.Error() == "email already exists" {
					http.Error(w, "Email already exists", http.StatusConflict)
					log.Println("Error signing up user:", err)
					return
				} else {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					log.Println("Error signing up user:", err)
					return
				}
			}
		case "Login":
			username, err := LogIn(w, database, r.FormValue("Email"), r.FormValue("Password"))
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				log.Println("Error logging in:", err)
				return
			}
			if username == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if err := CreateCookie(w, database, username); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Println("Error creating cookie:", err)
				return
			}
			fmt.Println("logged in as:", r.FormValue("Email"), "username:", username)
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func SignUp(w http.ResponseWriter, db *sql.DB, name, email, password string) error {
	exists, err := CheckEmail(db, email)
	if err != nil {
		return fmt.Errorf("error checking email: %v", err)
	}
	if exists {
		return fmt.Errorf("email already exists")
	}

	exists, err = CheckUserName(db, name)
	if err != nil {
		return fmt.Errorf("error checking username: %v", err)
	}
	if exists {
		return fmt.Errorf("username already exists")
	}

	if err := InsertUser(db, name, email, password); err != nil {
		return fmt.Errorf("error inserting user: %v", err)
	}
	return nil
}

func LogIn(w http.ResponseWriter, db *sql.DB, email, password string) (string, error) {
	username := ""
	exists, err := CheckEmail(db, email)
	if err != nil {
		return "", fmt.Errorf("error checking email: %v", err)
	}
	if !exists {
		return "", nil // User does not exist
	}

	stmt := `SELECT PASSWORD, USERNAME FROM USER WHERE EMAIL = ?`
	row := db.QueryRow(stmt, email)
	var dbPassword string
	err = row.Scan(&dbPassword, &username)
	if err != nil {
		return "", fmt.Errorf("error retrieving user data: %v", err)
	}

	if dbPassword != password {
		return "", nil // Incorrect password
	}

	return username, nil
}

func InsertUser(db *sql.DB, name, email, password string) error {
	userUUID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("error generating UUID: %v", err)
	}

	_, err = db.Exec("INSERT INTO USER(UUID, EMAIL, USERNAME, PASSWORD) VALUES(?, ?, ?, ?)",
		userUUID.String(), email, name, password)
	if err != nil {
		return fmt.Errorf("error inserting user: %v", err)
	}

	return nil
}

func CreateCookie(w http.ResponseWriter, db *sql.DB, username string) error {
	stmt := `SELECT UUID FROM USER WHERE USERNAME = ?`
	row := db.QueryRow(stmt, username)
	var userUUID string
	if err := row.Scan(&userUUID); err != nil {
		return fmt.Errorf("error retrieving user UUID: %v", err)
	}

	expiresAt := time.Now().Add(3600 * time.Second)
	newSession := session{
		UserUUID: userUUID,
		UserName: username,
		expiry:   expiresAt,
	}

	sessions = append(sessions, newSession)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    userUUID,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
	fmt.Println("cookie created")
	return nil
}

func CheckEmail(db *sql.DB, email string) (bool, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return false, nil // Invalid email format, treat as if it doesn't exist
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM USER WHERE EMAIL = ?", email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func CheckUserName(db *sql.DB, name string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM USER WHERE USERNAME = ?", name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
