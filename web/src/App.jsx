import { useEffect, useRef, useState } from 'react'

const GLOBAL_ROOM_ID = 'global'

function App() {
  const [connected, setConnected] = useState(false)
  const socketRef = useRef(null)

  const [rooms, setRooms] = useState([]) // browsable list: { roomID, roomName }
  const [joinedRoomIDs, setJoinedRoomIDs] = useState([GLOBAL_ROOM_ID])
  const [activeRoomID, setActiveRoomID] = useState(GLOBAL_ROOM_ID)
  const [messagesByRoom, setMessagesByRoom] = useState({}) // { roomID: [message] }

  const [draft, setDraft] = useState('')
  const [nameDraft, setNameDraft] = useState('')
  const [roomNameDraft, setRoomNameDraft] = useState('')

  const send = (payload) => {
    if (socketRef.current && connected) {
      socketRef.current.send(JSON.stringify(payload))
    }
  }

  // open the websocket once, on mount
  useEffect(() => {
    const socket = new WebSocket(`ws://${window.location.host}/ws`)
    socketRef.current = socket

    socket.onopen = () => {
      setConnected(true)
      socket.send(JSON.stringify({ type: 'listRooms' }))
    }
    socket.onclose = () => setConnected(false)

    socket.onmessage = (event) => {
      const incoming = JSON.parse(event.data)

      switch (incoming.type) {
        case 'message':
          setMessagesByRoom((previous) => ({
            ...previous,
            [incoming.roomID]: [...(previous[incoming.roomID] || []), incoming],
          }))
          break

        case 'roomList':
          setRooms(incoming.rooms || [])
          break

        case 'roomCreated':
        case 'roomJoined':
          setJoinedRoomIDs((previous) =>
            previous.includes(incoming.roomID)
              ? previous
              : [...previous, incoming.roomID]
          )
          setRooms((previous) =>
            previous.some((room) => room.roomID === incoming.roomID)
              ? previous
              : [...previous, { roomID: incoming.roomID, roomName: incoming.roomName }]
          )
          setActiveRoomID(incoming.roomID)
          break

        case 'roomLeft':
          setJoinedRoomIDs((previous) =>
            previous.filter((id) => id !== incoming.roomID)
          )
          setActiveRoomID((current) =>
            current === incoming.roomID ? GLOBAL_ROOM_ID : current
          )
          break

        case 'error':
          alert(incoming.text)
          break

        default:
          console.log('unhandled message from server:', incoming)
      }
    }

    return () => socket.close()
  }, [])

  const sendChat = () => {
    const text = draft.trim()
    if (text === '') return
    send({ type: 'message', roomID: activeRoomID, text })
    setDraft('')
  }

  const setName = () => {
    const name = nameDraft.trim()
    if (name === '') return
    send({ type: 'setName', displayName: name })
    setNameDraft('')
  }

  const createRoom = () => {
    const roomName = roomNameDraft.trim()
    if (roomName === '') return
    send({ type: 'createRoom', roomName })
    setRoomNameDraft('')
  }

  // already a member -> just view it; otherwise ask to join
  const openRoom = (room) => {
    if (joinedRoomIDs.includes(room.roomID)) {
      setActiveRoomID(room.roomID)
    } else {
      send({ type: 'joinRoom', roomID: room.roomID })
    }
  }

  const leaveRoom = (roomID) => {
    send({ type: 'leaveRoom', roomID })
  }

  const refreshRooms = () => {
    send({ type: 'listRooms' })
  }

  const activeMessages = messagesByRoom[activeRoomID] || []
  const activeRoom = rooms.find((room) => room.roomID === activeRoomID)
  const activeRoomName = activeRoom ? activeRoom.roomName : activeRoomID

  return (
    <div className="app-window">
      <div className="status-bar">
        status: {connected ? '🟢 connected' : '🔴 disconnected'}
      </div>

      <div className="toolbar">
        <div>
          <input
            value={nameDraft}
            onChange={(event) => setNameDraft(event.target.value)}
            placeholder="your display name"
          />
          <button onClick={setName}>Set name</button>
        </div>
        <div>
          <input
            value={roomNameDraft}
            onChange={(event) => setRoomNameDraft(event.target.value)}
            placeholder="new room name"
          />
          <button onClick={createRoom}>Create room</button>
        </div>
      </div>

      <div className="app-body">
        <div className="sidebar">
          <h3>
            ROOMS <button className="leave-button" onClick={refreshRooms}>refresh</button>
          </h3>
          {rooms.map((room) => {
            const joined = joinedRoomIDs.includes(room.roomID)
            const active = room.roomID === activeRoomID
            return (
              <div key={room.roomID}>
                <button
                  className={`room-entry${active ? ' active' : ''}`}
                  onClick={() => openRoom(room)}
                >
                  {room.roomName}
                  {joined ? '' : ' (join)'}
                </button>
                {joined && room.roomID !== GLOBAL_ROOM_ID && (
                  <button
                    className="leave-button"
                    onClick={() => leaveRoom(room.roomID)}
                  >
                    leave
                  </button>
                )}
              </div>
            )
          })}
        </div>

        <div className="chatbox">
          <h1 className="room-heading">{activeRoomName}</h1>
          <div className="message-window">
            {activeMessages.map((message, index) => (
              <div className="message-line" key={index}>
                <strong>{message.sender}:</strong> {message.text}
              </div>
            ))}
          </div>
          <div className="composer">
            <input
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              onKeyDown={(event) => event.key === 'Enter' && sendChat()}
              placeholder={`message ${activeRoomName}`}
            />
            <button onClick={sendChat}>Send</button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
