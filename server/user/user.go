// user
//
// wird verwendet um zu prüfen ob ein Nutzer angemeldet ist
// prüft dies über Cookies
//
// meldet Nutzer an und ab

package user

import (
	"net/http"
	"sync"
)

// Username ist der Name vom aktuell angemeldeten Nutzer
var Username string

// WG ist eine WaitGroup die genutzt wird um CheckUser Zeit zu geben die Cookies zu prüfen.
var WG sync.WaitGroup

// CheckUser prüft ob ein Nutzer angemeldet ist und gibt den Namen zurück
func CheckUser(r *http.Request) string {

	// holt Cookie mit dem Namen login
	cookie, _ := r.Cookie("login")
	Username := ""

	// wenn Cookie existiert wird der Name gespeichert
	if cookie != nil {
		Username = cookie.Value
	}

	// Überprüfung ist durch -> Main darf weitermachen
	defer WG.Done()

	// Nutzernamen zurückgeben
	return Username
}

// GetName gibt den Namen vom Nutzer zurück
func GetName() string {
	return Username
}

// LoggedOut meldet den Nutzer ab
func LoggedOut() {
	Username = ""
}

// LoggedIn meldet den Nutzer an
func LoggedIn(username string) {
	Username = username
}
