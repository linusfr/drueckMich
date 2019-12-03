//
// fängt die Nachricht von background ab
// schickt URL vom aktiven Tab als Nachricht an background
//

// Fängt Nachrichten ab
chrome.runtime.onMessage.addListener(function(request, sender, sendResponse) {
  // Wenn die Nachricht von background ist
  if (request.message === "fetchURL") {
    // URL vom aktiven Tab speichern
    var hrefActiveTab = window.location.href;

    // URL an Background schicken
    chrome.runtime.sendMessage({
      message: "URL",
      href: hrefActiveTab
    });
  }
});
