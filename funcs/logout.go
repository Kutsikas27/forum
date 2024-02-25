package funcs

import (
	"net/http"
	"time"
)

func LogOut(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	val := cookie.Value
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for index, session := range sessions {
		if session.UserUUID == val {
			sessions[index] = sessions[len(sessions)-1]
			sessions = sessions[:len(sessions)-1]
		}
	}
	cookie.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
