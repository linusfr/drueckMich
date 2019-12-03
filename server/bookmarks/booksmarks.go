// bookmarks
//
// Verwaltet alle Lesezeichen
//
// Erstellt neue Lesezeichen - löscht einzelne - löscht alle von einem Nutzer
//
// holt Informationen für Lesezeichen über Webscraper
//
// Alle Lesezeichen werden in einer Collection gespeichert
// Lesezeichen werden über das Feld Owner Nutzern zugeordnet

package bookmarks

import (
	"fmt"
	"strings"
	"sync"

	"../user"
	"../webscraper"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Collection für die Lesezeichen
var bookmarksCollection *mgo.Collection

// Init wird von der Main aufgerufen und schickt die collection für die Authentifikation
func Init(c *mgo.Collection) {
	bookmarksCollection = c
}

// WG wird genutzt um Aktionen die länger brauchen die Zeit dazu zu geben
var WG sync.WaitGroup

// Website ist der Struct für Websites
type Website struct {
	Owner       string `bson:"owner"`
	IconURL     string `bson:"iconurl"`
	Name        string `bson:"name"`
	URL         string `bson:"url"`
	Description string `bson:"description"`
}

// Add wird genutzt um neue Lesezeichen hinzuzufügen
func Add(url string, username string) {

	var website Website

	name, iconURL, description := webscraper.GetStructForURL(url)

	website.Owner = username
	website.Name = name
	website.IconURL = iconURL
	website.Description = description
	website.URL = url

	count, _ := bookmarksCollection.Find(bson.M{"url": url, "owner": username}).Count()

	if count == 0 {
		err := bookmarksCollection.Insert(website)

		if err != nil {
			fmt.Println("could NOT be inserted.")
			fmt.Println("ERROR:")
			fmt.Println(err)
		} else {
			fmt.Println("the url was inserted into the Bookmarks Collection")
			fmt.Println("")
			// fmt.Print(bookmarksCollection.Find(bson.M{}))
		}
	}
}

// Import wird genutzt um Chrome Lesezeichen zu speichern
func Import(website Website) {

	// schaut ob das Lesezeichen schon existiert
	count, _ := bookmarksCollection.Find(bson.M{"url": website.URL, "owner": website.Owner}).Count()

	// Wenn es noch nicht existiert
	if count == 0 {

		// Wenn die URL nicht leer ist und ein Eigentümer existiert
		if website.URL != "" && website.Owner != "" {

			// Speicher die Website ein
			err := bookmarksCollection.Insert(website)

			// Wenn die Website nicht eingespeichert werden konnte
			if err != nil {

				// gib Fehlermeldung aus
				fmt.Println("could NOT be inserted.")
				fmt.Println("ERROR:")
				fmt.Println(err)

			} else {

				// gib Erfolgsmeldung aus
				fmt.Println("the url was importet into the Bookmarks Collection")
				fmt.Println("")
			}

		}
	}
}

// GetUserWebsites holt alle Lesezeichen von einem Nutzer, sortiert diese ggf. und gibt sie dann aus als Array
func GetUserWebsites(sortBy string) []Website {

	// Array aus Websites für alle Lesezeichen
	websites := []Website{}

	// Sortiert nach Namen
	if sortBy == "sortNameDown" {
		bookmarksCollection.Find(bson.M{"owner": user.GetName()}).Sort("-name").All(&websites)
	}
	if sortBy == "sortNameUp" {
		bookmarksCollection.Find(bson.M{"owner": user.GetName()}).Sort("name").All(&websites)
	}

	// Sortiert nach URL
	if sortBy == "sortURLDown" {
		bookmarksCollection.Find(bson.M{"owner": user.GetName()}).Sort("-url").All(&websites)
	}
	if sortBy == "sortURLUp" {
		bookmarksCollection.Find(bson.M{"owner": user.GetName()}).Sort("url").All(&websites)
	}
	if sortBy == "" {
		bookmarksCollection.Find(bson.M{"owner": user.GetName()}).All(&websites)
	}

	return websites
}

// Delete löscht ein bestimmtes Lesezeichen von einem bestimmten Owner
func Delete(username string, url string) {

	// sucht Lesezeichen mit angegebener URL und owner und löscht dieses
	bookmarksCollection.Remove(bson.M{"url": url, "owner": username})
}

// DropOwner löscht alle Lesezeichen von einem Owner
func DropOwner(username string) {

	// das kann länger dauern wenn der Nutzer viele Lesezeichen hatte
	// deshalb wird vorm Aufruf der Funktion die WaitGroup erhöht damit auf diese Funktion gewartet wird

	// sucht alle Lesezeichen von einem Owner raus und löscht diese
	bookmarksCollection.Remove(bson.M{"owner": username})

	// Funktion ist durch -> WaitGroup reduzieren damit die Main weitermacht
	defer WG.Done()
}

// ImportChromeBookmarks bekommt die HTML Datei vom Chrome Export und pflegt die Lesezeichen daraus ein
func ImportChromeBookmarks(html string) {

	// Jedes Lesezeichen ist umgeben von DT und A Tags
	// Man bekommt die einzelnen Lesezeichen indem man nach dem Start Tag und vor dem End Tag splittet

	// Teilt die html Datei in seine Lesezeichen auf
	htmlArray := strings.Split(html, "<DT><A")

	var website Website
	username := user.GetName()

	// geht durch alle Lesezeichen und erstellt aus dem HTML Websites
	for _, item := range htmlArray {

		// owner setzen
		website.Owner = username

		// webscraper beauftragen aus dem string Informationen zu extrahieren
		url, name, description, iconURL := webscraper.CrawlImport(item)

		// Werte in Struct schreiben
		website.URL = url
		website.Name = name
		website.Description = description
		website.IconURL = iconURL

		// Wenn kein Name enthalten war und keiner durch den scraper geholt werden konnte default Wert setzen
		if website.Name == "" {
			website.Name = "Es ist kein Name angegeben."
		}

		// loggen
		fmt.Println("OWNER:")
		fmt.Println(website.Owner)

		// Import die fertige Website überreichen zum einpflegen in die Collection
		Import(website)
	}
}
