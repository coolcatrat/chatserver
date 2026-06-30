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

// ============
// ROOM STATE
// ============
const GLOBAL_ROOM_ID = "global";     // mirrors server's GlobalRoomID
let currentRoomID = GLOBAL_ROOM_ID; 
let ownDisplayName = null;            // set by server.

// ===========================================================================
// CONNECTION
// ===========================================================================
const webSocketProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
const webSocketConnection = new WebSocket(`${webSocketProtocol}//${window.location.host}/ws`);

webSocketConnection.onopen = function () {
  connectionStatusIndicator.textContent = "connected";
  promptForDisplayName();
};

webSocketConnection.onclose = function () {
  connectionStatusIndicator.textContent = "disconnected";
};

// ===========================================================================
// INCOMING  (server -> browser) has to mirror OutgoingMessage defined on serverside
// ===========================================================================
const incomingMessageHandlers = {
  message: displayChatMessage,// display new incoming message

};


webSocketConnection.onmessage = function (receivedEvent) {
  let parsedMessage;
  try{
    parsedMessage = JSON.parse(receivedEvent.data);
  }catch(parseError) {
    console.error("bad json from server: ",receivedEvent.data,parseError);
    return;
  }

  const messageHandler = incomingMessageHandlers[parsedMessage.type];
  if (messageHandler) messageHandler(parsedMessage);
  else console.warn("undefined message type: ", parsedMessage.type);

};

function displayChatMessage(chatMessage) {
  const messageLineElement = document.createElement("div");
  messageLineElement.className = "messageLine";

  const senderNameElement = document.createElement("span");
  senderNameElement.className = "senderName";
  senderNameElement.textContent= chatMessage.sender;
  senderNameElement.style.color = colorForSender(chatMessage.sender);

  const messageBodyNode = document.createTextNode(": "+ chatMessage.text);

  messageLineElement.appendChild(senderNameElement);
  messageLineElement.appendChild(messageBodyNode);

  messageDisplayWindow.appendChild(messageLineElement);
  messageDisplayWindow.scrollTop = messageDisplayWindow.scrollHeight; // keep newest in view
}

// ===========================================================================
// OUTGOING  (browser -> server)
// ===========================================================================

// universal function to send JSON to server.
function sendToServer(messageObject){
    webSocketConnection.send(JSON.stringify(messageObject));  // single send 
}

// protocol handler, wraps chatMessage into required JSON.
function sendChatMessage(messageText) {
  sendToServer({
    type: "message",
    roomID: currentRoomID,
    text: messageText,
  });
}

// reacts to "send". updates html, calls sendChatMessage()
function sendCurrentInput() {
  const messageText = messageInputField.value.trim();
  if (messageText === "") return;
  sendChatMessage(messageText);
  messageInputField.value = "";
}

function sendSetDisplayName(displayName) {
  sendToServer({ type: "setName", displayName: displayName });
}

// ===========
// display name 
// ===========



function promptForDisplayName() {
  const enteredName = window.prompt("Choose a display name:");
  if (enteredName === null) return;        // Cancel → keep server placeholder
  const trimmedName = enteredName.trim();
  if (trimmedName === "") return;          // empty → keep placeholder
  sendSetDisplayName(trimmedName);
  connectionStatusIndicator.textContent = "connected as " + ownDisplayName;

}


// ===========================================================================
// EVENTLISTENERS 
// ===========================================================================

// SENDING MESSAGE 
sendMessageButton.addEventListener("click",sendCurrentInput);
messageInputField.addEventListener("keydown", function(keyboardEvent){
    if(keyboardEvent.key === "Enter") sendCurrentInput();
});


// color 
function colorForSender(senderName) {
  let hashAccumulator = 0;
  for (let characterIndex = 0; characterIndex < senderName.length; characterIndex++) {
    const characterCode = senderName.charCodeAt(characterIndex);
    hashAccumulator = characterCode + ((hashAccumulator << 5) - hashAccumulator); // characterCode + hashAccumulator*31
  }
  const hue = Math.abs(hashAccumulator) % 360;  // fold into 0–359 colour-wheel degrees
  return `hsl(${hue}, 70%, 45%)`;               // fixed saturation/lightness → uniform readability
}


