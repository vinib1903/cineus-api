package ws

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/vinib1903/cineus-api/internal/domain/room"
	"github.com/vinib1903/cineus-api/internal/ports/http/httputil"
)

// Handler gerencia as conexões WebSocket.
type Handler struct {
	hub      *Hub
	roomRepo room.Repository
}

// NewHandler cria um novo handler WebSocket.
func NewHandler(hub *Hub, roomRepo room.Repository) *Handler {
	return &Handler{
		hub:      hub,
		roomRepo: roomRepo,
	}
}

// HandleConnection processa uma nova conexão WebSocket.
// GET /ws/room/{roomId}
func (h *Handler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket: new connection attempt from %s", r.RemoteAddr)

	// 1. Extrair o ID da sala da URL
	roomID := chi.URLParam(r, "roomId")
	log.Printf("WebSocket: roomId = %s", roomID)

	if roomID == "" {
		log.Println("WebSocket: room ID is empty")
		httputil.BadRequest(w, "Room ID is required")
		return
	}

	// 2. Extrair informações do usuário do contexto
	userID := httputil.GetUserID(r.Context())
	log.Printf("WebSocket: userID = %s", userID)

	if userID == "" {
		log.Println("WebSocket: user not authenticated")
		httputil.Unauthorized(w, "Authentication required")
		return
	}

	// 3. Buscar a sala no banco
	log.Printf("WebSocket: looking for room %s in database", roomID)
	rm, err := h.roomRepo.GetByID(r.Context(), room.ID(roomID))
	if err != nil {
		log.Printf("WebSocket: room %s not found: %v", roomID, err)
		httputil.NotFound(w, "Room not found")
		return
	}
	log.Printf("WebSocket: found room %s (%s)", rm.ID, rm.Name)

	// 4. Fazer o upgrade da conexão HTTP para WebSocket
	log.Println("WebSocket: attempting to accept connection...")
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("WebSocket: failed to accept connection: %v", err)
		return
	}
	log.Println("WebSocket: connection accepted!")

	// 5. Obter ou criar o RoomHub
	roomHub := h.hub.GetOrCreateRoom(RoomConfig{
		RoomID:    string(rm.ID),
		RoomName:  rm.Name,
		RoomTheme: string(rm.Theme),
		OwnerID:   string(rm.OwnerID),
		MaxSeats:  rm.MaxSeats,
	})

	// 6. Criar displayName temporário
	displayName := "User-" + userID[:8]

	// 7. Criar o cliente
	client := NewClient(roomHub, conn, userID, displayName)

	// 8. Registrar o cliente
	roomHub.register <- client

	log.Printf("WebSocket: user %s connected to room %s", userID, roomID)

	// 9. Iniciar (bloqueia até desconectar)
	client.Run()

	log.Printf("WebSocket: user %s disconnected from room %s", userID, roomID)
}

// GetStats retorna estatísticas do WebSocket.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]int{
		"rooms":   h.hub.GetRoomCount(),
		"clients": h.hub.GetTotalClients(),
	}
	httputil.JSON(w, http.StatusOK, stats)
}
