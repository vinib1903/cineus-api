package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RoomHub gerencia os clientes de uma sala.
type RoomHub struct {
	// ID da sala
	roomID string

	// Informações da sala
	roomName  string
	roomTheme string
	ownerID   string
	maxSeats  int

	// Clientes conectados: userID -> Client
	clients map[string]*Client

	// Assentos: seatID -> userID
	seats map[string]string

	// Estado do player de mídia
	mediaState *MediaState

	// Canais de comunicação
	register   chan *Client
	unregister chan *Client
	broadcast  chan *OutgoingMessage

	// Mutex para proteger maps e mediaState
	mu sync.RWMutex

	// Referência ao hub global
	globalHub *Hub
}

// NewRoomHub cria um novo hub de sala.
func NewRoomHub(globalHub *Hub, roomID, roomName, roomTheme, ownerID string, maxSeats int) *RoomHub {
	hub := &RoomHub{
		roomID:     roomID,
		roomName:   roomName,
		roomTheme:  roomTheme,
		ownerID:    ownerID,
		maxSeats:   maxSeats,
		clients:    make(map[string]*Client),
		seats:      make(map[string]string),
		mediaState: nil, // Sem vídeo inicialmente
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *OutgoingMessage, 256),
		globalHub:  globalHub,
	}

	// Inicializar assentos vazios
	for i := 1; i <= maxSeats; i++ {
		seatID := string(rune('A'-1+i/10+1)) + string(rune('0'+i%10))
		if i <= 9 {
			seatID = "A" + string(rune('0'+i))
		}
		hub.seats[seatID] = ""
	}

	return hub
}

// Run inicia o loop principal do hub.
func (h *RoomHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

// handleRegister adiciona um cliente à sala.
func (h *RoomHub) handleRegister(client *Client) {
	h.mu.Lock()

	// Verificar se usuário já está na sala
	if existingClient, exists := h.clients[client.userID]; exists {
		existingClient.Close()
	}

	h.clients[client.userID] = client
	h.mu.Unlock()

	log.Printf("Room %s: user %s joined (total: %d)", h.roomID, client.userID, len(h.clients))

	// Enviar estado inicial
	h.sendRoomState(client)

	// Notificar outros
	h.broadcastUserJoined(client)
}

// handleUnregister remove um cliente da sala.
func (h *RoomHub) handleUnregister(client *Client) {
	h.mu.Lock()

	if _, exists := h.clients[client.userID]; !exists {
		h.mu.Unlock()
		return
	}

	// Liberar assento
	seatID := client.GetSeatID()
	if seatID != "" {
		h.seats[seatID] = ""
	}

	delete(h.clients, client.userID)
	clientCount := len(h.clients)
	h.mu.Unlock()

	log.Printf("Room %s: user %s left (total: %d)", h.roomID, client.userID, clientCount)

	h.broadcastUserLeft(client.userID)

	if seatID != "" {
		h.broadcastSeatUpdated(seatID, nil)
	}

	if clientCount == 0 {
		log.Printf("Room %s: empty, removing from global hub", h.roomID)
		h.globalHub.removeRoom(h.roomID)
	}
}

// handleBroadcast envia mensagem para todos os clientes.
func (h *RoomHub) handleBroadcast(message *OutgoingMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		client.Send(message)
	}
}

// handleMessage processa uma mensagem recebida de um cliente.
func (h *RoomHub) handleMessage(client *Client, msg *IncomingMessage) {
	switch msg.Type {
	case TypeChatMessage:
		h.handleChatMessage(client, msg.Payload)

	case TypeSelectSeat:
		h.handleSelectSeat(client, msg.Payload)

	case TypeMediaControl:
		h.handleMediaControl(client, msg.Payload)

	default:
		client.SendError("UNKNOWN_TYPE", "Unknown message type")
	}
}

// handleChatMessage processa uma mensagem de chat.
func (h *RoomHub) handleChatMessage(client *Client, payload json.RawMessage) {
	var chatPayload ChatMessagePayload
	if err := json.Unmarshal(payload, &chatPayload); err != nil {
		client.SendError("INVALID_PAYLOAD", "Invalid chat message payload")
		return
	}

	if chatPayload.Content == "" {
		client.SendError("EMPTY_MESSAGE", "Message content cannot be empty")
		return
	}

	if len(chatPayload.Content) > 500 {
		client.SendError("MESSAGE_TOO_LONG", "Message cannot exceed 500 characters")
		return
	}

	broadcastPayload := ChatMessagePayload{
		ID:          uuid.New().String(),
		UserID:      client.userID,
		DisplayName: client.displayName,
		Content:     chatPayload.Content,
		CreatedAt:   time.Now(),
	}

	h.broadcast <- NewOutgoingMessage(TypeChatMessage, broadcastPayload)
}

// handleSelectSeat processa a seleção de assento.
func (h *RoomHub) handleSelectSeat(client *Client, payload json.RawMessage) {
	var seatPayload SelectSeatPayload
	if err := json.Unmarshal(payload, &seatPayload); err != nil {
		client.SendError("INVALID_PAYLOAD", "Invalid seat selection payload")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	currentOccupant, exists := h.seats[seatPayload.SeatID]
	if !exists {
		client.SendError("INVALID_SEAT", "Seat does not exist")
		return
	}

	if currentOccupant != "" && currentOccupant != client.userID {
		client.SendError("SEAT_OCCUPIED", "Seat is already occupied")
		return
	}

	oldSeatID := client.GetSeatID()
	if oldSeatID != "" && oldSeatID != seatPayload.SeatID {
		h.seats[oldSeatID] = ""
		go h.broadcastSeatUpdated(oldSeatID, nil)
	}

	h.seats[seatPayload.SeatID] = client.userID
	client.SetSeatID(seatPayload.SeatID)

	userID := client.userID
	go h.broadcastSeatUpdated(seatPayload.SeatID, &userID)
}

// handleMediaControl processa comandos de controle de mídia.
func (h *RoomHub) handleMediaControl(client *Client, payload json.RawMessage) {
	// Apenas o host pode controlar a mídia
	if client.userID != h.ownerID {
		client.SendError("NOT_HOST", "Only the room owner can control media")
		return
	}

	var controlPayload MediaControlPayload
	if err := json.Unmarshal(payload, &controlPayload); err != nil {
		client.SendError("INVALID_PAYLOAD", "Invalid media control payload")
		return
	}

	h.mu.Lock()

	// Inicializar mediaState se não existir
	if h.mediaState == nil {
		h.mediaState = &MediaState{
			VideoURL:    "",
			VideoTitle:  "",
			IsPlaying:   false,
			CurrentTime: 0,
			UpdatedAt:   time.Now(),
		}
	}

	switch controlPayload.Action {
	case MediaActionPlay:
		h.mediaState.IsPlaying = true
		h.mediaState.UpdatedAt = time.Now()

	case MediaActionPause:
		h.mediaState.IsPlaying = false
		h.mediaState.UpdatedAt = time.Now()

	case MediaActionSeek:
		if controlPayload.Time < 0 {
			h.mu.Unlock()
			client.SendError("INVALID_TIME", "Time cannot be negative")
			return
		}
		h.mediaState.CurrentTime = controlPayload.Time
		h.mediaState.UpdatedAt = time.Now()

	case MediaActionChange:
		if controlPayload.VideoURL == "" {
			h.mu.Unlock()
			client.SendError("INVALID_URL", "Video URL is required")
			return
		}
		h.mediaState.VideoURL = controlPayload.VideoURL
		h.mediaState.VideoTitle = controlPayload.VideoTitle
		h.mediaState.CurrentTime = 0
		h.mediaState.IsPlaying = true
		h.mediaState.UpdatedAt = time.Now()

	default:
		h.mu.Unlock()
		client.SendError("INVALID_ACTION", "Invalid media control action")
		return
	}

	// Copiar estado para broadcast
	stateCopy := *h.mediaState
	h.mu.Unlock()

	log.Printf("Room %s: media %s by %s", h.roomID, controlPayload.Action, client.userID)

	// Broadcast do novo estado para todos
	h.broadcast <- NewOutgoingMessage(TypeMediaState, MediaStatePayload{
		Media:     stateCopy,
		UpdatedBy: client.userID,
	})
}

// sendRoomState envia o estado atual da sala para um cliente.
func (h *RoomHub) sendRoomState(client *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomInfo := RoomInfo{
		ID:       h.roomID,
		Name:     h.roomName,
		Theme:    h.roomTheme,
		OwnerID:  h.ownerID,
		MaxSeats: h.maxSeats,
	}

	users := make([]UserInfo, 0, len(h.clients))
	for _, c := range h.clients {
		users = append(users, UserInfo{
			ID:          c.userID,
			DisplayName: c.displayName,
			SeatID:      c.GetSeatID(),
		})
	}

	seats := make([]SeatInfo, 0, len(h.seats))
	i := 0
	for seatID, userID := range h.seats {
		seat := SeatInfo{
			ID:       seatID,
			Position: i,
		}
		if userID != "" {
			seat.UserID = &userID
		}
		seats = append(seats, seat)
		i++
	}

	// Incluir estado da mídia se existir
	var mediaState *MediaState
	if h.mediaState != nil {
		stateCopy := *h.mediaState
		mediaState = &stateCopy
	}

	client.Send(NewOutgoingMessage(TypeRoomState, RoomStatePayload{
		Room:  roomInfo,
		Users: users,
		Seats: seats,
		Media: mediaState,
	}))
}

// broadcastUserJoined notifica que um usuário entrou.
func (h *RoomHub) broadcastUserJoined(client *Client) {
	msg := NewOutgoingMessage(TypeUserJoined, UserJoinedPayload{
		User: UserInfo{
			ID:          client.userID,
			DisplayName: client.displayName,
		},
	})

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, c := range h.clients {
		if c.userID != client.userID {
			c.Send(msg)
		}
	}
}

// broadcastUserLeft notifica que um usuário saiu.
func (h *RoomHub) broadcastUserLeft(userID string) {
	h.broadcast <- NewOutgoingMessage(TypeUserLeft, UserLeftPayload{
		UserID: userID,
	})
}

// broadcastSeatUpdated notifica mudança de assento.
func (h *RoomHub) broadcastSeatUpdated(seatID string, userID *string) {
	h.broadcast <- NewOutgoingMessage(TypeSeatUpdated, SeatUpdatedPayload{
		SeatID: seatID,
		UserID: userID,
	})
}

// GetOwnerID retorna o ID do dono da sala.
func (h *RoomHub) GetOwnerID() string {
	return h.ownerID
}
