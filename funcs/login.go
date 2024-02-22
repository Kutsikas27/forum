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

var sessions = map[string]session{}

type session struct {
	UserUUID string
	UserName string
	expiry   time.Time
}

// cookie expiery time
func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	email, name, password := "", "", ""

	if r.URL.Path != "/login" {
		HandleErrorPage(w, r)
		return
	}

	if r.Method == "POST" {
		database, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Fatal("Error opening database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer database.Close()
		if r.FormValue("operation") == "signup" {
			name = r.FormValue("Name")
			password = r.FormValue("Password")
			email = r.FormValue("Email")
			Credentials(w, database, email, name, password)
			createCookie(w, database, name)
		} else if r.FormValue("operation") == "Login" {
			password = r.FormValue("Password")
			email = r.FormValue("Email")
			user, err := login(w, database, email, password)
			if err != nil {
				log.Println("Error logging in:", err)
				return
			}
			if user == "" {
				http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
				return
			}
			createCookie(w, database, user)
		}

	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// checks if the user cridentials are valid
func Credentials(w http.ResponseWriter, database *sql.DB, email, name, password string) {
	exists1, err := checkEmail(database, email)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	exists2, err := checkUserName(database, name)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if exists1 || exists2 {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	err = insertUser(database, email, name, password)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

}

// checks if email is aleady in use
func checkEmail(db *sql.DB, email string) (bool, error) {
	valid := validateEmail(email)
	if !valid {
		return true, nil
	}
	stmt := `SELECT	EMAIL FROM USER WHERE EMAIL = ?`
	err := db.QueryRow(stmt, email).Scan(&email)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// checks if the email is valid
func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// checks if the username already in use
func checkUserName(db *sql.DB, name string) (bool, error) {
	stmt := `SELECT	USERNAME FROM USER WHERE USERNAME = ?`
	err := db.QueryRow(stmt, name).Scan(&name)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// puts newly created user into the database
func insertUser(db *sql.DB, email, name, password string) error {
	useruuid, err := uuid.NewV4()
	if err != nil {
		return err
	}
	strUuid := useruuid.String()
	insert := `INSERT INTO USER(UUID, EMAIL, USERNAME, PASSWORD) VALUES(?, ?, ?, ?)`
	statement, err := db.Prepare(insert)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer statement.Close()

	_, err = statement.Exec(strUuid, email, name, password)
	if err != nil {
		return fmt.Errorf("error executing statement: %v", err)
	}

	return nil
}

// creates cookie to save userdata
func createCookie(w http.ResponseWriter, db *sql.DB, username string) {
	stmt := `SELECT	UUID FROM USER WHERE USERNAME = ?`
	row := db.QueryRow(stmt, username)
	var useruuid string
	err := row.Scan(&useruuid)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	cookieUUID, err := uuid.NewV4()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tokenString := cookieUUID.String()

	expiresAt := time.Now().Add(3600 * time.Second)
	sessions[tokenString] = session{
		UserUUID: useruuid,
		UserName: username,
		expiry:   expiresAt,
	}

	http.SetCookie(w, &http.Cookie{
		Name:       "session_token",
		Value:      tokenString,
		Path:       "/",
		Domain:     "",
		Expires:    expiresAt,
		RawExpires: "",
		MaxAge:     3600,
		Secure:     true,
		HttpOnly:   true,
		SameSite:   0,
		Raw:        "",
		Unparsed:   []string{},
	})
	fmt.Println("cookie created")
}

func login(w http.ResponseWriter, database *sql.DB, email, password string) (string, error) {
	Username := ""
	exists, err := checkEmail(database, email)
	if err != nil {
		log.Fatal("Error checking credentials:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return "", err
	}
	if !exists {
		log.Fatal("Error email doesnt exist:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return "", err
	}
	stmt := `SELECT	PASSWORD FROM USER WHERE EMAIL = ?`
	row := database.QueryRow(stmt, email)
	var dbPassword string
	err = row.Scan(&dbPassword)
	if err != nil {
		log.Fatal("Error scaning database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return "", err
	}

	if dbPassword == password {
		stmt := `SELECT	USERNAME FROM USER WHERE EMAIL = ?`
		row := database.QueryRow(stmt, email)
		var names string
		err = row.Scan(&names)
		if err != nil {
			log.Fatal("Error scaning database:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return "", err
		}
		Username = names
	}
	return Username, nil
}
