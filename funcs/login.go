package funcs

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"text/template"

	"github.com/gofrs/uuid/v5"
	_ "github.com/mattn/go-sqlite3"
)

var logintmp = template.Must(template.New("signin.html").ParseFiles("frontend/templates/signin.html"))

func LoginPage(w http.ResponseWriter, r *http.Request) {
	email, name, password := "", "", ""

	if r.URL.Path != "/signin" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		err := logintmp.Execute(w, nil)
		if err != nil {
			log.Fatal("Error executing template:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

	} else if r.Method == "POST" {
		name = r.FormValue("Name")
		password = r.FormValue("Password")
		email = r.FormValue("Email")
		Credentials(w, email, name, password)
		http.Redirect(w, r, "/", http.StatusFound)
	}

}

func Credentials(w http.ResponseWriter, email, name, password string) {
	database, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	exists1, err := checkEmail(database, email)
	if err != nil {
		log.Fatal("Error checking credentials:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	exists2, err := checkUserName(database, name)
	if err != nil {
		log.Fatal("Error checking credentials:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if exists1 || exists2 {
		log.Fatal("Error username or email already exists")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	err = insertUser(database, email, name, password)
	if err != nil {
		log.Fatal("Error adding user to database")
	}

}

func checkEmail(db *sql.DB, email string) (bool, error) {
	valid := validateEmail(email)
	fmt.Println(email)
	fmt.Println(valid)
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

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

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
