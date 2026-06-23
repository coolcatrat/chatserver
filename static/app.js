// ===========================================================================
// HANDLES
// ===========================================================================
const connectionStatusIndicator = document.getElementById("connectionStatusIndicator");
const messageDisplayWindow      = document.getElementById("messageDisplayWindow");
const messageInputField         = document.getElementById("messageInputField");
const sendMessageButton         = document.getElementById("sendMessageButton");
const currentRoomHeading        = document.getElementById("currentRoomHeading");
const globalChatRoomList        = document.getElementById("globalChatRoomList");
const directMessageList         = document.getElementById("directMessageList");

// ===========================================================================
// CONNECTION
// ===========================================================================
const webSocketProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
const webSocketConnection = new WebSocket(`${webSocketProtocol}//${window.location.host}/ws`);

webSocketConnection.onopen = function () {
  connectionStatusIndicator.textContent = "connected";
};

webSocketConnection.onclose = function () {
  connectionStatusIndicator.textContent = "disconnected";
};

// ===========================================================================
// INCOMING  (server -> browser)
// ===========================================================================
webSocketConnection.onmessage = function (receivedEvent) {
  const messageLineElement = document.createElement("div");
  messageLineElement.className = "messageLine";     
  messageLineElement.textContent = receivedEvent.data;
  messageDisplayWindow.appendChild(messageLineElement);  
};

// ===========================================================================
// OUTGOING  (browser -> server)
// ===========================================================================
function sendCurrentInput() {
  const message = messageInputField.value;
  if(message == "") return;
  webSocketConnection.send(message);
  messageInputField.value = "";
}


// ===========================================================================
// EVENTLISTENERS 
// ===========================================================================

// SENDING MESSAGE 
sendMessageButton.addEventListener("click",sendCurrentInput);
messageInputField.addEventListener("keydown", function(keyboardEvent){
    if(keyboardEvent.key === "Enter") sendCurrentInput();
});



