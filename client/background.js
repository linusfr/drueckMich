//
// Wenn der Action Button geklickt wird fragt background beim content die URL an
// Wenn die Antwort von content kommt -> URL an server schicken und client öffnen
//

document.addEventListener("DOMContentLoaded", function() {
  // action Button wurde geklickt
  chrome.browserAction.onClicked.addListener(function(tab) {
    // Sendet Nachricht an Content die URL zu holen
    chrome.tabs.query(
      {
        active: true,
        currentWindow: true
      },
      function(tabs) {
        var activeTab = tabs[0];
        chrome.tabs.sendMessage(activeTab.id, {
          message: "fetchURL"
        });
      }
    );
  });

  // fängt Antworten ab
  chrome.runtime.onMessage.addListener(function(request, sender, sendResponse) {
    // Wenn die Antwort von Content ist
    if (request.message === "URL") {
      var value = {
        href: request.href
      };
      // nimm die URL und mach einen String draus
      var valueAsJSON = JSON.stringify(value);

      // Url für den Request
      var postURL = "http://localhost:4242/postURL";
      var postURLRequest = new XMLHttpRequest();

      // Ajax Request wird an den Server geschickt mit der URL
      postURLRequest.open("POST", postURL + "?valueAsJSON=" + valueAsJSON);
      postURLRequest.setRequestHeader("Content-type", "application/json");

      // Anfrage abschicken
      postURLRequest.send();

      // Antwort abfangen
      postURLRequest.addEventListener("load", function() {
        // öffnet den Client in einem neuen Tab
        // Wenn schon offen -> hinspringen und neu laden

        var newTabURL = "http://localhost:4242/client";

        // sucht nach einem Tab mit der URL vom client
        chrome.tabs.query({ url: newTabURL }, function(tabs) {
          // Wen ein Tab mit dem Client existiert
          if (tabs[0]) {
            // gehe zum Tab
            chrome.tabs.update(tabs[0].id, { active: true });
            // lade ihn neu
            chrome.tabs.reload(tabs[0].id);
          } else {
            // client existiert nicht -> neuen Tab erstellen und client öffnen
            chrome.tabs.create({ url: newTabURL });
          }
        });
      });
    }
  });
});
