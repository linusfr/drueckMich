// Webscraper dient dazu URLs nach Informationen zu durchsuchen
// Dazu werden Anfragen an die URLs geschickt um den Source Code zu bekommen
// Dieser wird dann durchsucht
//

package webscraper

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// PageURL ist die URL die durchsucht wird
var PageURL string

// GetStructForURL durchsucht eine gegebene URL nach Informationen und gibt diese zurück
func GetStructForURL(url string) (string, string, string) {

	// setzt die URL als globale damit die anderen Funktionen damit arbeiten können
	PageURL = url

	// ruft die entsprechenden Funktionen auf um die Werte auszulesen
	iconURL := GetJPG()
	name := GetName()
	description := GetDescription()

	// log
	fmt.Println("Name:")
	fmt.Println(name)
	fmt.Println("Icon URL:")
	fmt.Println(iconURL)
	fmt.Println("URL:")
	fmt.Println(PageURL)
	fmt.Println("Description:")
	fmt.Println(description)

	return name, iconURL, description
}

// CrawlImport holt die Informationen für ein importiertes Lesezeichen
func CrawlImport(item string) (string, string, string, string) {

	// holt Informationen
	url, name := getInfo(item)

	// setzt URL für die anderen Funktionen
	PageURL = url

	var description string
	var iconURL string

	// Wenn die URL existiert stell get Request
	if url != "" {
		// Holt Informationen
		description = GetDescription()
		iconURL = GetJPG()
	}

	// Falls kein Name enthalten war durchsuche den SourceCode nach einem
	if name == "" {
		name = GetName()
	}

	// logg
	fmt.Println("Name:")
	fmt.Println(name)
	fmt.Println("URL:")
	fmt.Println(url)
	fmt.Println("Decription:")
	fmt.Println(description)
	fmt.Println("iconURL:")
	fmt.Println(iconURL)

	return url, name, description, iconURL
}

// extrahiert Informationen aus einem importierten Lesezeichen
func getInfo(item string) (string, string) {

	// Format von item das die Funktion bekommt
	// HREF="http://theLink.de">linkText

	// ------------- URL ---------------------

	var temp string
	// nach HREF splitten um alles vor URL Beginn zu verwerfen
	array := strings.Split(item, `HREF="`)
	for i := range array {
		if i == 1 {
			temp = array[i]
		}
	}

	// Neues Format
	// http://theLink.de">linkText

	// Alles Nach Linkende verwerfen
	array = strings.SplitN(temp, `"`, 2)
	temp = array[0]

	// Neues Format
	// http://theLink.de

	// URl speichern
	url := temp

	// ---------- Name -----------------

	// Alles vor > verwerfen um nur den Namen zu haben
	array = strings.Split(item, ">")
	for i := range array {
		if i == 1 {
			temp = array[i]
		}
	}

	array = strings.Split(temp, "</A")
	name := array[0]

	// Werte zurückgeben
	return url, name
}

// GetJPG returns the URL of an ICON for the set URL
func GetJPG() string {

	// geht in 3 Schritten vor um Bild zu bekommen. Wenn ein Schritt scheitert wird er nächste probiert
	// 1. -> guckt nach einem Link mit einer Relation von Icon
	// 2. falls nicht gefunden -> guckt nach einem Icon in irgendeinem Link
	// 3. immernoch nicht gefunden -> sucht nach irgendeinem Bild

	// Variable für Icon URL
	var linkToIcon string

	// falls ein Bild gefunden wird werden alle nächsten Schritte übersprungen
	found := false

	// Schickt get Request für die URL
	res, err := http.Get(GetRootURL())

	// Abfangen falls nicht erfolgreichs
	if err != nil {
		fmt.Println(err)
		linkToIcon = ""

	} else {

		//---------------------------------------------------------------------------------------------------
		// 1. -> guckt nach einem Link mit einer Relation von Icon
		//---------------------------------------------------------------------------------------------------

		// erstellt einen Tokenizer aus dem Response Body
		page := html.NewTokenizer(res.Body)

		for !found {
			_ = page.Next()
			token := page.Token()

			// wenn kein Token übrig ist aus der Schleife springen
			if token.Type == html.ErrorToken {
				break
			}

			// prüft ob der aktuelle Token ein Link ist
			if token.DataAtom == atom.Link {

				// prüft ob der Token rel"=icon" enthält und ein jpg oder png ist
				isIconAndJPG := strings.Contains(token.String(), `rel="icon"`) && strings.Contains(token.String(), ".jpg")
				isIconAndPNG := strings.Contains(token.String(), `rel="icon"`) && strings.Contains(token.String(), ".png")

				if isIconAndJPG || isIconAndPNG {

					// Icon wurde gefunden -> springt zum Ende
					found = true

					// URL zum Icon extrahieren

					// nach href splitten um alles vor linkbeginn zu verwerfen
					array := strings.Split(token.String(), `href="`)
					for i := range array {
						if i == 1 {
							linkToIcon = array[1]
						}
					}

					// alles nach Ende vom Link verwerfen
					array = strings.Split(linkToIcon, `"`)
					linkToIcon = array[0]

					// falls der URL link ein relativer ist einen absoluten daraus erstellen
					if !strings.Contains(linkToIcon, "//") {
						linkToIcon = GetRootURL() + linkToIcon
					}
				}
			}
		}

		//---------------------------------------------------------------------------------------------------
		// 2. falls nicht gefunden -> sucht nach irgendeinem Link mit Icon drin
		//---------------------------------------------------------------------------------------------------

		// Schickt get Request für die URL
		res, _ = http.Get(GetRootURL())

		// erstellt einen Tokenizer aus dem Response Body
		page = html.NewTokenizer(res.Body)

		for !found {
			_ = page.Next()
			token := page.Token()

			// wenn kein Token übrig ist aus der Schleife springen
			if token.Type == html.ErrorToken {
				break
			}

			// prüfen ob der Link ein Token ist
			if token.DataAtom == atom.Link {

				// prüft ob der Link ein jpg oder png enthält
				linkContainsPNG := strings.Contains(token.String(), ".jpg")
				linkContainsJPG := strings.Contains(token.String(), ".png")

				if linkContainsPNG || linkContainsJPG {

					// Icon gefunden -> Spring zum Ende
					found = true

					// URL extrahieren

					// nach href splitten um alles vor urlbeginn zu verwerfen
					array := strings.Split(token.String(), `href="`)
					for i := range array {
						if i == 1 {
							linkToIcon = array[1]
						}
					}

					// alles nach urlende verwerfen
					array = strings.Split(linkToIcon, `"`)
					linkToIcon = array[0]

					// wenn es ein relativer Link ist einen absoluten daraus machen
					if !strings.Contains(linkToIcon, "//") {
						linkToIcon = GetRootURL() + linkToIcon
					}
				}
			}
		}

		//---------------------------------------------------------------------------------------------------
		// 3. immernoch nicht gefunden -> nach irgendeinem Icon suchen
		//---------------------------------------------------------------------------------------------------

		// Schickt get Request für die URL
		res, _ = http.Get(GetRootURL())

		// erstellt einen Tokenizer aus dem Response Body
		page = html.NewTokenizer(res.Body)

		for !found {

			_ = page.Next()
			token := page.Token()

			// wenn kein Token übrig ist aus der Schleife springen
			if token.Type == html.ErrorToken {
				break
			}

			// prüft ob der Token ein Icon enthält
			isPNG := strings.Contains(token.String(), ".png")
			isJPG := strings.Contains(token.String(), ".jpg")

			if isPNG || isJPG {

				// if -> icon ist in einem href
				// elseif -> icon ist in einem content
				// else -> weder noch -> hol das erste Icon

				iconInHref := strings.Contains(token.String(), `href="`)
				iconInContent := strings.Contains(token.String(), `content="`)

				if iconInHref {

					// Icon gefunden -> spring zum Ende
					found = true

					// nach href splitten um alles vor Url beginn zu verwerfen
					array := strings.Split(token.String(), `href="`)
					for i := range array {
						if i == 1 {
							linkToIcon = array[1]
						}
					}

					// alles nach url ende verwerfen
					array = strings.Split(linkToIcon, `"`)
					linkToIcon = array[0]

				} else if iconInContent {

					// icon gefunden -> spring zum Ende
					found = true

					// nach href splitten um alles vor Url beginn zu verwerfen
					array := strings.Split(token.String(), `content="`)
					for i := range array {
						if i == 1 {
							linkToIcon = array[1]
						}
					}

					// alles nach url ende verwerfen
					array = strings.Split(linkToIcon, `"`)
					linkToIcon = array[0]

				} else {

					// icon gefunden -> spring zum Ende
					found = true

					// nach =" splitten um alles vor Url beginn zu verwerfen
					array := strings.Split(token.String(), `="`)
					for i := range array {
						if i == 1 {
							linkToIcon = array[1]
						}
					}

					// alles nach url ende verwerfen
					array = strings.Split(linkToIcon, `"`)
					linkToIcon = array[0]

				}

				// falls url relativ ist absolut machen
				if !strings.Contains(linkToIcon, "//") {
					linkToIcon = GetRootURL() + linkToIcon
				}
			}
		}
	}
	return linkToIcon

}

// GetName extrahiert den Namen aus einer URL
func GetName() string {

	var name string

	// Schickt get Request für die URL
	res, error := http.Get(PageURL)

	if error != nil {
		fmt.Println(error)
		name = "Es ist kein Name angegeben."
	} else {

		// erstellt einen Tokenizer aus dem Response Body
		page := html.NewTokenizer(res.Body)

		found := false

		for !found {
			_ = page.Next()
			token := page.Token()

			if token.Type == html.ErrorToken {
				break
			}

			// prüft ob der Token ein Titel ist
			if token.DataAtom == atom.Title {
				_ = page.Next()

				// Inhalt vom nächsten Token speichern -> Inhalt vom Titel
				token := page.Token()
				name = token.String()

				// gefunden -> zum Ende Springen
				found = true

				// Kein Titel gefunden -> Default Wert speichern
			} else {
				name = "Es ist kein Name angegeben."
			}
		}
	}

	return name
}

// GetDescription sucht die Beschreibung raus
func GetDescription() string {

	var description string

	// Schickt get Request für die URL
	res, err := http.Get(PageURL)

	// Default Wert speichern falls Anfrage nicht möglich
	if err != nil {
		fmt.Println(err)
		description = "Es ist keine Beschreibung angegeben."
	} else {

		// erstellt einen Tokenizer aus dem Response Body
		page := html.NewTokenizer(res.Body)

		found := false
		for !found {
			_ = page.Next()
			token := page.Token()

			if token.Type == html.ErrorToken {
				break
			}

			// prüft ob der Token ein Meta Tag ist
			if token.DataAtom == atom.Meta {

				// prüft ob der Meta Tag eine Beschreibung enthält
				if strings.Contains(token.String(), `name="Description"`) {

					// Alles vor Beginn der Beschreibung verwerfen
					printArray := strings.SplitAfter(token.String(), `content="`)

					// Alles nach Ende der Beschreibung verwerfen
					printArray = strings.Split(printArray[1], `">`)
					description = printArray[0]

					// Gefunden -> Zum Ende Springen
					found = true

				} else {

					// Default Wert setzen
					description = "Es ist keine Beschreibung angegeben."
				}
			}
		}
	}
	return description
}

// GetRootURL extrahiert die Root URL aus einer URL
func GetRootURL() string {

	// Alles nach / verwerfen
	urlArray := strings.Split(PageURL, "/")

	// Altes Format:
	// https://blablabla.bla/blabla
	// Alles nach drittem / verwerfen
	// wird zu
	// https://blablabla.bla

	// die ersten beiden / wieder einsetzen und Wert zurückgeben
	return urlArray[0] + "//" + urlArray[1] + urlArray[2]

}
