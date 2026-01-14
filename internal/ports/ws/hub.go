package ws

import (
	"log"
	"sync"
)

// Hub é o gerenciador global de todas as salas.
type Hub struct {
	// Mapa de salas: roomID -> RoomHub
	rooms map[string]*RoomHub

	// Mutex para proteger o mapa
	mu sync.RWMutex
}

// NewHub cria um novo hub global.
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*RoomHub),
	}
}

// RoomConfig contém as configurações para criar uma sala.
type RoomConfig struct {
	RoomID    string
	RoomName  string
	RoomTheme string
	OwnerID   string
	MaxSeats  int
}

// GetOrCreateRoom retorna uma sala existente ou cria uma nova.
func (h *Hub) GetOrCreateRoom(cfg RoomConfig) *RoomHub {
	// Primeiro, tenta ler com RLock (mais rápido)
	h.mu.RLock()
	if room, exists := h.rooms[cfg.RoomID]; exists {
		h.mu.RUnlock()
		return room
	}
	h.mu.RUnlock()

	// Não existe, precisa criar com Lock completo
	h.mu.Lock()
	defer h.mu.Unlock()

	// Verificar novamente (outro goroutine pode ter criado)
	if room, exists := h.rooms[cfg.RoomID]; exists {
		return room
	}

	// Criar nova sala
	room := NewRoomHub(h, cfg.RoomID, cfg.RoomName, cfg.RoomTheme, cfg.OwnerID, cfg.MaxSeats)
	h.rooms[cfg.RoomID] = room

	// Iniciar o loop da sala em uma goroutine
	go room.Run()

	log.Printf("Hub: created room %s (%s)", cfg.RoomID, cfg.RoomName)

	return room
}

// GetRoom retorna uma sala existente ou nil se não existir.
func (h *Hub) GetRoom(roomID string) *RoomHub {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID]
}

// removeRoom remove uma sala do hub (chamado quando a sala fica vazia).
func (h *Hub) removeRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.rooms, roomID)
	log.Printf("Hub: removed room %s", roomID)
}

// GetRoomCount retorna o número de salas ativas.
func (h *Hub) GetRoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

// GetTotalClients retorna o número total de clientes conectados.
func (h *Hub) GetTotalClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, room := range h.rooms {
		room.mu.RLock()
		total += len(room.clients)
		room.mu.RUnlock()
	}
	return total
}

