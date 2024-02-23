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

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	email, name, password := "", "", ""

	if r.URL.Path != "/login" {
		HandleErrorPage(w, r)
		return
	}

	if r.Method == "POST" {
		database, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Fatal("Error opening database:", err)
			return
		}
		defer database.Close()
		if r.FormValue("operation") == "signup" {
			name = r.FormValue("Name")
			password = r.FormValue("Password")
			email = r.FormValue("Email")
			err := Credentials(w, database, email, name, password)
			if err != nil {
				log.Fatal("Error creating user:", err)
				return
			}
			err = createCookie(w, database, name)
			if err != nil {
				log.Fatal("Error creating cookie:", err)
				return
			}
		} else if r.FormValue("operation") == "Login" {
			password = r.FormValue("Password")
			email = r.FormValue("Email")
			user, err := login(w, database, email, password)
			if err != nil {
				log.Fatal("Error logging in:", err)
				return
			}
			if user == "" {
				http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
				return
			}
			err = createCookie(w, database, user)
			if err != nil {
				log.Fatal("Error creating cookie:", err)
				return
			}
			fmt.Println("logged in as:", email, "username:", user, "password", password)
		}

	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// checks if the user cridentials are valid
func Credentials(w http.ResponseWriter, database *sql.DB, email, name, password string) error {
	exists1, err := checkEmail(database, email)
	if err != nil {
		log.Println("Error checking email:", err)
		return err
	}

	exists2, err := checkUserName(database, name)
	if err != nil {
		log.Println("Error checking username:", err)
		return err
	}
	if exists1 || exists2 {
		return fmt.Errorf("email or username already exists")
	}

	err = insertUser(database, email, name, password)
	if err != nil {
		return err
	}
	return nil
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
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(strUuid, email, name, password)
	if err != nil {
		return err
	}

	return nil
}

// creates cookie to save userdata
func createCookie(w http.ResponseWriter, db *sql.DB, username string) error {
	stmt := `SELECT	UUID FROM USER WHERE USERNAME = ?`
	row := db.QueryRow(stmt, username)
	var useruuid string
	err := row.Scan(&useruuid)
	if err != nil {
		return err
	}
	cookieUUID, err := uuid.NewV4()
	if err != nil {
		return err
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
	return nil
}

func login(w http.ResponseWriter, database *sql.DB, email, password string) (string, error) {
	Username := ""
	exists, err := checkEmail(database, email)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}
	stmt := `SELECT	PASSWORD FROM USER WHERE EMAIL = ?`
	row := database.QueryRow(stmt, email)
	var dbPassword string
	err = row.Scan(&dbPassword)
	if err != nil {
		return "", err
	}

	if dbPassword == password {
		stmt := `SELECT	USERNAME FROM USER WHERE EMAIL = ?`
		row := database.QueryRow(stmt, email)
		var names string
		err = row.Scan(&names)
		if err != nil {
			return "", err
		}
		Username = names
	} else {
		return "", nil
	}
	return Username, nil
}
