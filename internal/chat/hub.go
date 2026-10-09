package chat

import (
	"sync"
	"time"
)

// MessageType distinguishes user messages from system announcements.
type MessageType int

const (
	MsgUser MessageType = iota
	MsgSystem
	MsgJoin
	MsgLeave
)

// Message represents a live lobby transmission.
type Message struct {
	ID        int64       `json:"id"`
	Type      MessageType `json:"type"`
	Author    string      `json:"author"`
	Content   string      `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
	ColorHex  string      `json:"color_hex"`
}

// UserInfo represents active visitor presence data.
type UserInfo struct {
	SessionID string
	Nickname  string
	Status    string
	Room      string
	JoinedAt  time.Time
}

// Client represents an active connected session subscribing to chat.
type Client struct {
	SessionID string
	Nickname  string
	Status    string
	Room      string
	Muted     bool
	SendCh    chan Message
	JoinedAt  time.Time
}

// Hub coordinates multiplayer chat and presence across all SSH sessions.
type Hub struct {
	mu          sync.RWMutex
	clients     map[string]*Client
	history     []Message
	maxHistory  int
	msgSeq      int64
	broadcastCh chan Message
	registerCh  chan *Client
	unregCh     chan string
}

// NewHub initializes a new multiplayer Chat Hub.
func NewHub(maxHistory int) *Hub {
	if maxHistory <= 0 {
		maxHistory = 100
	}
	h := &Hub{
		clients:     make(map[string]*Client),
		history:     make([]Message, 0, maxHistory),
		maxHistory:  maxHistory,
		broadcastCh: make(chan Message, 128),
		registerCh:  make(chan *Client, 32),
		unregCh:     make(chan string, 32),
	}

	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.registerCh:
			h.mu.Lock()
			h.clients[client.SessionID] = client
			h.mu.Unlock()

			// Broadcast arrival to others
			h.BroadcastSystem(client.Nickname + " materialized in the network.")

		case sessionID := <-h.unregCh:
			h.mu.Lock()
			client, exists := h.clients[sessionID]
			if exists {
				delete(h.clients, sessionID)
				close(client.SendCh)
			}
			h.mu.Unlock()

			if exists {
				h.BroadcastSystem(client.Nickname + " faded back into the dark.")
			}

		case msg := <-h.broadcastCh:
			h.mu.Lock()
			// Append to history
			if len(h.history) >= h.maxHistory {
				h.history = h.history[1:]
			}
			h.history = append(h.history, msg)

			// Distribute to all connected clients
			for _, client := range h.clients {
				if client.Muted && msg.Type == MsgUser {
					continue
				}
				select {
				case client.SendCh <- msg:
				default:
					// Dropped if client buffer is full to prevent head-of-line blocking
				}
			}
			h.mu.Unlock()
		}
	}
}

// Register registers a new client session.
func (h *Hub) Register(sessionID, nickname, status, room string) *Client {
	client := &Client{
		SessionID: sessionID,
		Nickname:  nickname,
		Status:    status,
		Room:      room,
		Muted:     false,
		SendCh:    make(chan Message, 64),
		JoinedAt:  time.Now(),
	}
	h.registerCh <- client
	return client
}

// Unregister disconnects a client session.
func (h *Hub) Unregister(sessionID string) {
	h.unregCh <- sessionID
}

// BroadcastUser sends a chat message from a visitor.
func (h *Hub) BroadcastUser(author, content, colorHex string) {
	h.mu.Lock()
	h.msgSeq++
	msg := Message{
		ID:        h.msgSeq,
		Type:      MsgUser,
		Author:    author,
		Content:   content,
		Timestamp: time.Now(),
		ColorHex:  colorHex,
	}
	h.mu.Unlock()
	h.broadcastCh <- msg
}

// BroadcastSystem sends a system event broadcast.
func (h *Hub) BroadcastSystem(content string) {
	h.mu.Lock()
	h.msgSeq++
	msg := Message{
		ID:        h.msgSeq,
		Type:      MsgSystem,
		Author:    "SYSTEM",
		Content:   content,
		Timestamp: time.Now(),
	}
	h.mu.Unlock()
	h.broadcastCh <- msg
}

// UpdatePresence updates a user's nickname and status.
func (h *Hub) UpdatePresence(sessionID, nickname, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[sessionID]; ok {
		c.Nickname = nickname
		c.Status = status
	}
}

// UpdateRoom updates which room the user is currently exploring.
func (h *Hub) UpdateRoom(sessionID, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[sessionID]; ok {
		c.Room = room
	}
}

// SetMute toggles mute state for a client.
func (h *Hub) SetMute(sessionID string, muted bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[sessionID]; ok {
		c.Muted = muted
	}
}

// GetRecentHistory returns a slice of recent messages.
func (h *Hub) GetRecentHistory() []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()
	msgs := make([]Message, len(h.history))
	copy(msgs, h.history)
	return msgs
}

// GetOnlineUsers returns a list of all active users.
func (h *Hub) GetOnlineUsers() []UserInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]UserInfo, 0, len(h.clients))
	for _, c := range h.clients {
		users = append(users, UserInfo{
			SessionID: c.SessionID,
			Nickname:  c.Nickname,
			Status:    c.Status,
			Room:      c.Room,
			JoinedAt:  c.JoinedAt,
		})
	}
	return users
}

// OnlineCount returns total active online users.
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// RoomDistribution returns count of users in each room.
func (h *Hub) RoomDistribution() map[string]int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	dist := make(map[string]int)
	for _, c := range h.clients {
		room := c.Room
		if room == "" {
			room = "Hub"
		}
		dist[room]++
	}
	return dist
}
