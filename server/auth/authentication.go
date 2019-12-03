// Authentication.go
//
// Zuständig für die Verwaltung der Benutzer
// Anmeldung, Registrierung etc.
//
// Bekommt eigene Collection von der Main exklusiv für die Nutzer
//

package auth

import (
	"errors"
	"fmt"

	"../bookmarks"
	"../user"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// definiert die Mindestlänge für Nutzernamen und Passwörter
const minPassLen = 2
const minNameLen = 3

// Collection für die Authentifikation
var authenticationCollection *mgo.Collection

// User ist der Struct der für Benutzer verwendet wird
type User struct {
	Name     string `bson:"name"`
	Password string `bson:"password"`
}

// Init wird von der Main aufgerufen und schickt die Collection zuständig für Benutzer
func Init(c *mgo.Collection) {
	authenticationCollection = c
}

// Register wird genutzt um neue Nutzer zu registrieren
func Register(username string, password string) error {
	// Wenn kein Nutzername eingegeben wurde
	if username == "" {
		return errors.New("Kein Benutzername eingegeben")
	}

	// Wenn kein Passwort angegeben wurde
	if password == "" {
		return errors.New("Kein Kennwort eingegeben")
	}

	// Wenn der Nutzername nicht lang genug ist
	if len(username) < minNameLen {
		return errors.New("Benutzername < 3 Zeichen")
	}

	// Wenn das Passwort nicht lang genug ist
	if len(password) < minPassLen {
		return errors.New("Kennwort < 3 Zeichen")
	}

	// guckt ob bereits ein Nutzer mit dem Benutzernamen existiert
	n, err := authenticationCollection.Find(bson.M{"name": username}).Count()

	if err != nil {
		fmt.Println(err)
	}

	// Falls bereits der Benutzername genutzt wird
	if n != 0 {
		return errors.New("Benutzername existiert bereits")
	}

	// Wenn alle Angaben korrekt waren und der Benutzername nicht verwendet wird
	// -> neuen Nutzer anlegen
	user := User{username, password}
	authenticationCollection.Insert(user)

	return nil
}

// DeleteAccount wird aufgerufen um Nutzer zu löschen
func DeleteAccount(username string) {

	// sucht den Nutzer mit dem angegebenen Namen raus und löscht ihn aus der Collection
	authenticationCollection.Remove(bson.M{"name": username})

	// weist bookmarks an alle Lesezeichen von dem Nutzer zu löschen
	bookmarks.DropOwner(username)

	// Loggt den Nutzer aus -> Username = ""
	user.LoggedOut()
}

// Login wird genutzt um einen Nutzer einzuloggen
func Login(username string, password string) error {

	// erstellt Dokument für einen Nutzer
	userDocument := User{}

	// prüft ob der Nutzer existiert und speichert ihn in userDocument zwischen
	err := authenticationCollection.Find(bson.M{"name": username}).One(&userDocument)

	// Wenn der Nutzer nicht existiert
	if err != nil {
		return errors.New("Benutzer nicht registriert")
	}

	// Prüft ob das Passwort korrekt ist
	if userDocument.Password != password {
		return errors.New("Kennwort falsch")
	}

	// gibt User bescheid das jetzt ein Nutzer angemeldet ist
	user.LoggedIn(username)

	return nil
}
