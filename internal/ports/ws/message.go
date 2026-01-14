package ws

import (
	"encoding/json"
	"time"
)

// MessageType define os tipos de mensagens WebSocket.
type MessageType string

const (
	// Servidor → Cliente
	TypeRoomState   MessageType = "room_state"
	TypeUserJoined  MessageType = "user_joined"
	TypeUserLeft    MessageType = "user_left"
	TypeSeatUpdated MessageType = "seat_updated"
	TypeMediaState  MessageType = "media_state"
	TypeMediaSync   MessageType = "media_sync"
	TypeError       MessageType = "error"

	// Cliente → Servidor
	TypeChatMessage  MessageType = "chat_message"
	TypeSelectSeat   MessageType = "select_seat"
	TypeMediaControl MessageType = "media_control"
	TypeAvatarAction MessageType = "avatar_action"
)

// IncomingMessage é a estrutura de mensagens recebidas do cliente.
type IncomingMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// OutgoingMessage é a estrutura de mensagens enviadas para o cliente.
type OutgoingMessage struct {
	Type      MessageType `json:"type"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewOutgoingMessage cria uma nova mensagem de saída.
func NewOutgoingMessage(msgType MessageType, payload interface{}) *OutgoingMessage {
	return &OutgoingMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// --- Payloads específicos ---

// RoomStatePayload é o estado inicial da sala.
type RoomStatePayload struct {
	Room  RoomInfo    `json:"room"`
	Users []UserInfo  `json:"users"`
	Seats []SeatInfo  `json:"seats"`
	Media *MediaState `json:"media,omitempty"`
}

// RoomInfo são informações básicas da sala.
type RoomInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Theme    string `json:"theme"`
	OwnerID  string `json:"owner_id"`
	MaxSeats int    `json:"max_seats"`
}

// UserInfo são informações de um usuário na sala.
type UserInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	SeatID      string `json:"seat_id,omitempty"`
}

// SeatInfo são informações de um assento.
type SeatInfo struct {
	ID       string  `json:"id"`
	Position int     `json:"position"`
	UserID   *string `json:"user_id,omitempty"`
}

// UserJoinedPayload é enviado quando alguém entra.
type UserJoinedPayload struct {
	User UserInfo `json:"user"`
}

// UserLeftPayload é enviado quando alguém sai.
type UserLeftPayload struct {
	UserID string `json:"user_id"`
}

// ChatMessagePayload é uma mensagem de chat.
type ChatMessagePayload struct {
	ID          string    `json:"id,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

// SelectSeatPayload é para escolher assento.
type SelectSeatPayload struct {
	SeatID string `json:"seat_id"`
}

// SeatUpdatedPayload é enviado quando um assento muda.
type SeatUpdatedPayload struct {
	SeatID string  `json:"seat_id"`
	UserID *string `json:"user_id"` // nil = assento liberado
}

// ErrorPayload é enviado quando ocorre um erro.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// --- Media Payloads ---

// MediaState representa o estado atual do player.
type MediaState struct {
	VideoURL    string    `json:"video_url"`
	VideoTitle  string    `json:"video_title"`
	IsPlaying   bool      `json:"is_playing"`
	CurrentTime float64   `json:"current_time"` // Em segundos
	UpdatedAt   time.Time `json:"updated_at"`
}

// MediaStatePayload é enviado quando o estado do player muda.
type MediaStatePayload struct {
	Media     MediaState `json:"media"`
	UpdatedBy string     `json:"updated_by"` // ID do usuário que fez a ação
}

// MediaControlAction define as ações possíveis de controle.
type MediaControlAction string

const (
	MediaActionPlay   MediaControlAction = "play"
	MediaActionPause  MediaControlAction = "pause"
	MediaActionSeek   MediaControlAction = "seek"
	MediaActionChange MediaControlAction = "change" // Trocar vídeo
)

// MediaControlPayload é enviado pelo host para controlar o player.
type MediaControlPayload struct {
	Action     MediaControlAction `json:"action"`
	Time       float64            `json:"time,omitempty"`        // Para seek
	VideoURL   string             `json:"video_url,omitempty"`   // Para change
	VideoTitle string             `json:"video_title,omitempty"` // Para change
}

// MediaSyncPayload é enviado periodicamente para manter sync.
type MediaSyncPayload struct {
	CurrentTime float64   `json:"current_time"`
	IsPlaying   bool      `json:"is_playing"`
	ServerTime  time.Time `json:"server_time"`
}
