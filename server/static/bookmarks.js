// bookmarks.js
//
// Requests:
// GET -> holt alle Website Dokumente vom Server und fügt sie dem Body der Tabelle hinzu
// POST -> Löscht das Element aus der Collection und aus der Tabelle
//
// Up/Down
// Hängt den Sortierpfeilen Eventlistener an die dafür sorgen,
// dass die Dokumente von der Collection auf dem Server nach Kategorie sortiert werden
// Fragt neu sortierte Dokumente an
//

window.addEventListener("load", () => {
  // -------------------------------- Sortieren -------------------------------------------------

  // URL für Serverfragen auf die der entsprechende Handler in Go registriert ist
  var url = "http://localhost:4242/updateBookmarks";

  // Variabel für die Anfrage
  var xhr = new XMLHttpRequest();

  // Initial wird eine Anfrage gestellt
  xhr.open("GET", url);
  xhr.send();

  // wonach wird sortiert?
  var sortName = true;
  var sortURL = true;

  // Event Handler für den Button zuständig für den Namen
  document.getElementById("sortName").addEventListener("click", function() {
    if (sortName) {
      // setzt Pfeil nach unten ein
      document.getElementById("sortNameIcon").src = "/img/down.png";

      // schickt Anfrage Liste sortiert nach Namen zu bekommen.
      xhr.open("GET", url + "?sortName=down");
      xhr.send();
    } else {
      // setzt Pfeil nach oben ein
      document.getElementById("sortNameIcon").src = "/img/up.png";

      // schickt Anfrage Liste sortiert nach Namen zu bekommen.
      xhr.open("GET", url + "?sortName=up");
      xhr.send();
    }

    // kehrt den Boolean um
    sortName = !sortName;
  });

  // Event Handler für den Button zuständig für die URL
  document.getElementById("sortURL").addEventListener("click", function() {
    if (sortURL) {
      // setzt Pfeil nach unten ein
      document.getElementById("sortURLIcon").src = "/img/down.png";

      // schickt Anfrage Liste sortiert nach URL zu bekommen.
      xhr.open("GET", url + "?sortURL=down");
      xhr.send();
    } else {
      // setzt Pfeil nach oben
      document.getElementById("sortURLIcon").src = "/img/up.png";

      // schickt Anfrage Liste sortiert nach URL zu bekommen.
      xhr.open("GET", url + "?sortURL=up");
      xhr.send();
    }

    // kehrt den Boolean um
    sortURL = !sortURL;
  });

  // --------------------------------- Interval --------------------------------------------------

  // Variabel für den Interval -> wird kontrolliert über set und clearTheInterval()
  // nutzt update und interval als parameter
  // -> alle 5 Sekunden wird vom Server ein neuer Body für die Tabelle angefragt
  var intervalHandle;

  // schickt Anfrag an den Server
  function update() {
    xhr.open("GET", url);
    xhr.send();
  }

  var interval = 5000; // 5 sec

  // spans bekommt alle span Elemente die es gibt -> Spans werden nur für den Löschen Button verwendet
  var spans;

  // jedes mal wenn eine Serverantwort kommt wird der Inhalt in den Body von der Tabelle eingefügt
  xhr.addEventListener("load", function() {
    document.getElementById("tableBody").innerHTML = xhr.responseText;

    // sucht alle Span Elemente und packt sie in spans
    spans = document.getElementsByTagName("span");
    // createListener hängt jedem span aus Spans einen EventListener an
    createListener();
  });

  // startet einen Interval in der Variabel intervalHandle
  setTheInterval();

  // setzt den Interval mit update und 5s Frequenz
  function setTheInterval() {
    intervalHandle = setInterval(update, interval);
  }

  // löscht den Interval
  function clearTheInterval() {
    clearInterval(intervalHandle);
  }

  // ------------------------------- Löschen von Tabellenzeilen ----------------------------------------

  // Erstellt einen EventListener für jedes Item aus spans.
  // -> Alle Lösch Buttons
  function createListener() {
    for (let item of spans) {
      item.addEventListener("click", function(event) {
        // Holt das DOM Element, dass das Event ausgelöst hat
        currentNode = event.currentTarget;

        // Struktur von allen span Knoten
        // <tr>
        //    <td>
        //      <span class="makeClickable"> <i class="fas fa-trash-alt"></i></span>
        //    </td>
        // </tr>

        // Wir wollen erst zur Tabellen Reihe und diese speichern um sie löschen zu können
        // dann wollen wir zum Element mit der URL vom Element um damit eine Löschanfrage an den Server schicken zu können.
        // span < td < tr > td > a.href

        // Geht zu TR -> Tabellenreihe
        currentNode = currentNode.parentNode.parentNode;

        // Zwischenspeichern um später löschen zu können
        tableRow = currentNode;

        // Jetzt wird die URL extrahiert für die Anfrage

        // Sucht die Löschspalte aus der Reihe raus
        tableChildrenArray = tableRow.childNodes;
        for (let item of tableChildrenArray) {
          if (item.className == "linkColumn") {
            currentNode = item;
            break;
          }
        }

        // Zieht das innerHTML aus dem Knoten
        goalURL = currentNode.innerHTML;

        // innerHTML von dem Knoten sieht so aus von der Struktur
        //	<a target="_blank" href="https://www.ipchicken.com/">
        //	<i class="fas fa-external-link-alt"></i></a>

        // Extrahieren von der URL

        // alles vor URL Beginn absplitten
        arr = goalURL.split(`href="`);
        goalURL = arr[1];

        // alles nach URL Ende absplitten
        arr = goalURL.split(`">`);
        goalURL = arr[0];

        // löschanfrage an den Server schicken
        var deletionRequest = new XMLHttpRequest();
        deletionRequest.open(
          "POST",
          "http://localhost:4242/updateBookmarks?URL=" + goalURL
        );
        deletionRequest.send();

        // Tabellenreihe löschen
        tableRow.remove();

        // Interval löschen damit der Server einen Augenblick Zeit hat die Anfrage zu verarbeiten
        clearTheInterval();

        // Sobald der Server bereit ist den Interval neu setzen.
        deletionRequest.addEventListener("load", function() {
          setTheInterval();
        });
      });
    }
  }
});
