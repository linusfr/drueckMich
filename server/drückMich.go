// main
//
// handelt alle Anfragen
// stellt File Server
// schickt die Templates
//
// Initialisiert Collections für Authentication und Bookmarks
//

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	// http Server
	"net/http"

	// html Templates
	"html/template"

	// Packages
	"./auth"
	"./bookmarks"
	"./user"

	// BSON
	"gopkg.in/mgo.v2"
)

// holt die Templates // Client - Import - Bookmarks - Tablecontent - login - logout
var t = template.Must(template.ParseFiles("template/client.html", "template/import.html", "template/bookmarks.html", "template/table.html", "template/login.html", "template/loggedOff.html"))

// deklariert Datenbank
var db *mgo.Database

// nach was wird sortiert?
var sortBy = ""

// Erhält URLs und schickt sie zu bookmarks wenn ein Nutzer angemeldet ist
func postURLHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	// liest die URL aus
	URL := r.FormValue("valueAsJSON")

	// altes Format:
	// {"href":"http://localhost:4242/client"}

	// Nach " splitten und den Dritten Teil nehmen
	urlArray := strings.Split(URL, `"`)
	URL = urlArray[3]

	// Neues Format:
	// http://localhost:4242/client

	// log
	fmt.Println("")
	fmt.Println(URL)

	// Wenn ein Nutzer angemeldet ist die URL in seiner Collection speichern

	// WG um sicherzustellen das der Nutzer geprüft wird
	user.WG.Add(1)

	// Holt Nutzernamen und prüft Cookie
	username := user.CheckUser(r)
	user.WG.Wait()

	// Nutzer angemeldet -> URL an Bookmarks schicken
	if username != "" {

		fmt.Println("send to bookmarks")

		bookmarks.Add(URL, username)

	} else {

		// kein Nutzer angemeldet
		// URL verwerfen
		fmt.Println("NOT send to bookmarks")
	}
}

// schickt bookmarks template -> bookmarks fragt über js alle 5 Sekunden nach Content
// -> siehe updateBookmarks
func bookmarkPageHandler(w http.ResponseWriter, r *http.Request) {

	// WG um sicherzustellen das der Nutzer geprüft wird
	user.WG.Add(1)

	// Holt Nutzernamen und prüft Cookie
	username := user.CheckUser(r)
	user.WG.Wait()

	// Nutzer nicht angemeldet -> Anmeldeseite schicken
	if username == "" {

		t.ExecuteTemplate(w, "login.html", nil)

	} else {

		// Nutzer angemeldet -> template schicken zusammen mit Nutzernamen für Navigationsleiste
		t.ExecuteTemplate(w, "bookmarks.html", username)
	}
}

// Website Array für updateBookmarks
var websites []bookmarks.Website

// Wird von Ajax Request von bookmarks.js aufgerufen um entweder alle Websites für den Nutzer anzufragen oder ein Lesezeichen zu löschen
func updateBookmarksPageHandler(w http.ResponseWriter, r *http.Request) {

	// WG um sicherzustellen das der Nutzer geprüft wird
	user.WG.Add(1)

	// Holt Nutzernamen und prüft Cookie
	username := user.CheckUser(r)
	user.WG.Wait()

	// Wenn Nutzer nicht angemeldet ist Anmeldeseite schicken
	if username == "" {
		t.ExecuteTemplate(w, "login.html", nil)
	} else {

		// GET wird genutzt um alle Lesezeichen ggf. sortiert anzufragen
		if r.Method == "GET" {
			r.ParseForm()

			// setzen wonach sortiert wird
			if r.FormValue("sortName") == "down" {
				sortBy = "sortNameDown"
			}
			if r.FormValue("sortName") == "up" {
				sortBy = "sortNameUp"
			}
			if r.FormValue("sortURL") == "down" {
				sortBy = "sortURLDown"
			}
			if r.FormValue("sortURL") == "up" {
				sortBy = "sortURLUp"
			}

			// websites holen und ggf. vorher sortieren lassen
			websites := bookmarks.GetUserWebsites(sortBy)

			// HTML Tabellenreihe für jedes Lesezeichen generieren und abschicken
			t.ExecuteTemplate(w, "table.html", websites)

			// genutzt um Anfragen zum Löschen eines Lesezeichens zu stellen
		} else if r.Method == "POST" {

			r.ParseForm()

			// URL auslesen
			URL := r.FormValue("URL")

			// log
			fmt.Println(URL)

			// Dokument aus der Collection löschen
			bookmarks.Delete(username, URL)
		}
	}
}

// Löscht den Account indem der Cookie zurückgesetzt wird und der Nutzer aus der Datenbank gelöscht wird
func acountDeletionHandler(w http.ResponseWriter, r *http.Request) {

	// WG um sicherzustellen das der Nutzer geprüft wird
	user.WG.Add(1)

	// Holt Nutzernamen und prüft Cookie
	username := user.CheckUser(r)
	user.WG.Wait()

	// Nutzer nicht angemeldet -> Anmeldeseite schicken
	if username == "" {

		t.ExecuteTemplate(w, "login.html", nil)

	} else {

		// WG um sicherzustellen das alle Lesezeichen gelöscht werden
		bookmarks.WG.Add(1)

		// Nutzer und all seine Lesezeichen löschen
		auth.DeleteAccount(username)

		// Weitermachen wenn alles gelöscht ist
		bookmarks.WG.Wait()

		// Cookie löschen
		cookie := http.Cookie{Name: "login", MaxAge: -1}
		http.SetCookie(w, &cookie)

		// template schicken mit Bestätigungsnachricht
		t.ExecuteTemplate(w, "loggedOff.html", "Dein Account wurde gelöscht.")
	}
}

// sign User out by deleting the cookie
func signOutHandler(w http.ResponseWriter, r *http.Request) {

	// WG um sicherzustellen das der Nutzer geprüft wird
	user.WG.Add(1)

	// Holt Nutzernamen und prüft Cookie
	username := user.CheckUser(r)
	user.WG.Wait()

	// Nutzer nicht angemeldet -> Anmeldeseite schicken
	if username == "" {

		t.ExecuteTemplate(w, "login.html", nil)

	} else {

		// Nutzer abmelden
		user.LoggedOut()

		// Cookie löschen
		cookie := http.Cookie{Name: "login", MaxAge: -1}
		http.SetCookie(w, &cookie)

		// template schicken mit Bestätigungsnachricht
		t.ExecuteTemplate(w, "loggedOff.html", "Du bist jetzt abgemeldet.")
	}
}

// Struct für Ausgabenachrichten vom Importhandler
type output struct {
	Name   string
	Output string
}

// Handler zum hinzufügen von exportierten Lesezeichen -> Wird von import.html verwendet
func importHandler(w http.ResponseWriter, r *http.Request) {

	// WG um sicherzustellen das der Nutzer geprüft wird
	user.WG.Add(1)

	// Holt Nutzernamen und prüft Cookie
	username := user.CheckUser(r)
	user.WG.Wait()

	// Nutzer nicht angemeldet -> Anmeldeseite schicken
	if username == "" {

		t.ExecuteTemplate(w, "login.html", nil)

	} else {

		// Fragt import Seite an
		if r.Method == "GET" {

			output := output{Name: username, Output: ""}

			// template schicken mit Nutzernamen für Navigationsleiste
			t.ExecuteTemplate(w, "import.html", output)

			// Genutzt um Datei zu schicken
		} else if r.Method == "POST" {

			// liest HTML File aus und speichert als byteArray
			byteArrayPage, err := ioutil.ReadAll(r.Body)

			// Erfolgreich?
			if err != nil {
				log.Fatal(err)
			}

			r.Body.Close()

			if err != nil {
				log.Fatal(err)
			}

			// byteArray zu String formatieren
			htmlAsString := (fmt.Sprintf("%s", byteArrayPage))

			// Header verwerfen
			arr := strings.Split(htmlAsString, "<H1>Bookmarks</H1>")

			bookmarks.ImportChromeBookmarks(arr[1])

			output := output{Name: username, Output: "Die Lesezeichen wurden erfolgreich importiert."}

			// template schicken mit Namen und Bestätigung das die Lesezeichen hinzugefügt wurden
			t.ExecuteTemplate(w, "import.html", output)
		}
	}
}

// zuständig für die Clientseite
func clientHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	// Fragt Seite an
	case "GET":

		// Cookie holen
		cookie, _ := r.Cookie("login")

		// Kein Nutzer angemeldet -> Anmeldeseite schicken
		if cookie == nil {

			t.ExecuteTemplate(w, "login.html", nil)

		} else {

			// Benutzer existiert
			// WG um sicherzustellen das der Nutzer angemeldet wird
			user.WG.Add(1)

			// Prüft Cookie und setzt Nutzernamen in Package
			user.CheckUser(r)
			user.WG.Wait()

			// schickt Client mit Nutzernamen
			t.ExecuteTemplate(w, "client.html", cookie.Value)
		}

	// genutzt von Login.html -> Genutzt zum Anmelden / Registrieren
	case "POST":
		r.ParseForm()

		// holt die FormValues
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Anmelden oder Registrieren?
		userWantsToLogin := r.FormValue("login") != ""
		userWantsToRegister := r.FormValue("register") != ""

		if userWantsToLogin {

			// Anmeldedaten überprüfen
			err := auth.Login(username, password)

			// Nutzer existiert nicht -> Anmeldeseite mit Fehlermeldung schicken
			if err != nil {

				fmt.Println(err)
				t.ExecuteTemplate(w, "login.html", err)

			} else {

				// Nutzer existiert -> erstell Cookie
				newCookie := http.Cookie{Name: "login", Value: username}
				http.SetCookie(w, &newCookie)

				// schickt client mit Nutzernamen
				t.ExecuteTemplate(w, "client.html", username)
				user.LoggedIn(username)
			}

		} else if userWantsToRegister {

			// Registrieungsdaten prüfen
			err := auth.Register(username, password)

			// Nutzernamen schon in Verwendung || zu kurz || passwort zu kurz
			if err != nil {

				// Loginseite mit Fehlermeldung schicken
				t.ExecuteTemplate(w, "login.html", err)

				// Nutzer konnte Registriert werden
			} else {

				// Erfolgreich registriert -> Anmeldeseite mit Bestätigung schicken
				t.ExecuteTemplate(w, "login.html", "Benutzer registriert")

			}
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// initialisiert Datenbank
// Registriert Handler
// Startet File Server
func main() {

	// -------------------------- Datenbank ---------------------------------

	// initialisiert Mongo
	session, _ := mgo.Dial("localhost") // mongodb://borsti.inf.fh-flensburg.de:27017
	defer session.Close()

	// Öffnet Datenbank
	db = session.DB("HA18DB_Linus_Frotscher_630063")

	// Schickt Collection zu Auth und Bookmarks
	auth.Init(db.C("auth"))
	bookmarks.Init(db.C("bookmarks"))

	// -------------------------- Fileserver ---------------------------------

	// File Server mit CSS und JS for the Templates
	http.Handle("/", http.FileServer(http.Dir("./static/")))

	// -------------------------- Handler ---------------------------------

	// Genutzt um URLs zu erhalten
	http.HandleFunc("/postURL", postURLHandler) // http://localhost:4242/postURL

	// Zum Anmelden oder um den Client zu bekommen
	http.HandleFunc("/client", clientHandler) // http://localhost:4242/client

	// Lesezeichenseite
	http.HandleFunc("/bookmark", bookmarkPageHandler) // http://localhost:4242/bookmark

	// Fragt über bookmark.js die Lesezeichen als HTML Tabellenreihen an
	http.HandleFunc("/updateBookmarks", updateBookmarksPageHandler) // http://localhost:4242/updateBookmarks

	// Zum Importieren von Chrome Lesezeichen
	http.HandleFunc("/import", importHandler) // http://localhost:4242/import

	// Abmelden
	http.HandleFunc("/signOut", signOutHandler) // http://localhost:4242/signOut

	// Account löschen
	http.HandleFunc("/deleteAccount", acountDeletionHandler) // http://localhost:4242/deleteAccount

	// port für den Server
	err := http.ListenAndServe(":4242", nil)

	// schiefgegangen?
	if err != nil {
		fmt.Println(err)
	}
}
