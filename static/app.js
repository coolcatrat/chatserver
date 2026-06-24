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
  let parsedMessage;
  try{
    parsedMessage = JSON.parse(receivedEvent.data);
  }catch(parseError) {
    console.error("bad json from server: ",receivedEvent.data,parseError);
    return;
  }

  switch(parsedMessage.type){
    case "message":
      displayChatMessage(parsedMessage);
      break;
    default:
      console.warn("unkown message type: ", parsedMessage.type);
  }
};

function displayChatMessage(chatMessage) {
  const messageLineElement = document.createElement("div");
  messageLineElement.className = "messageLine";
  messageLineElement.textContent = chatMessage.sender + ": " + chatMessage.text;
  messageDisplayWindow.appendChild(messageLineElement);
  messageDisplayWindow.scrollTop = messageDisplayWindow.scrollHeight; // keep newest in view
}
// ===========================================================================
// OUTGOING  (browser -> server)
// ===========================================================================
function sendCurrentInput() {
  const messageText = messageInputField.value.trim();
  if(messageText === "") return;

  const outgoingMessage = {
    type: "message",
    text: messageText,
  }

  webSocketConnection.send(JSON.stringify(outgoingMessage));
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



