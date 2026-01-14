package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	approom "github.com/vinib1903/cineus-api/internal/app/room"
	"github.com/vinib1903/cineus-api/internal/domain/room"
	"github.com/vinib1903/cineus-api/internal/domain/user"
	"github.com/vinib1903/cineus-api/internal/ports/http/httputil"
)

// RoomHandler gerencia as rotas de salas.
type RoomHandler struct {
	roomService *approom.Service
}

// NewRoomHandler cria uma nova instância do handler.
func NewRoomHandler(roomService *approom.Service) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

// RoomResponse é a representação de uma sala na resposta.
type RoomResponse struct {
	ID         string  `json:"id"`
	OwnerID    string  `json:"owner_id"`
	Name       string  `json:"name"`
	Theme      string  `json:"theme"`
	Visibility string  `json:"visibility"`
	AccessCode *string `json:"access_code,omitempty"` // Só retorna para o dono
	MaxSeats   int     `json:"max_seats"`
	CreatedAt  string  `json:"created_at"`
}

// toRoomResponse converte uma Room para RoomResponse.
func toRoomResponse(r *room.Room, includeCode bool) RoomResponse {
	resp := RoomResponse{
		ID:         string(r.ID),
		OwnerID:    string(r.OwnerID),
		Name:       r.Name,
		Theme:      string(r.Theme),
		Visibility: string(r.Visibility),
		MaxSeats:   r.MaxSeats,
		CreatedAt:  r.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Só inclui o código se for o dono
	if includeCode && r.AccessCode != nil {
		resp.AccessCode = r.AccessCode
	}

	return resp
}

// CreateRequest é o corpo da requisição de criação de sala.
type CreateRequest struct {
	Name       string `json:"name"`
	Theme      string `json:"theme"`
	Visibility string `json:"visibility"`
}

// Create cria uma nova sala.
// POST /api/v1/rooms
func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Extrair o ID do usuário do contexto
	userID := httputil.GetUserID(r.Context())
	if userID == "" {
		httputil.Unauthorized(w, "User not authenticated")
		return
	}

	// Decodificar o corpo
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Invalid request body")
		return
	}

	// Validar campos
	if req.Name == "" {
		httputil.BadRequest(w, "Name is required")
		return
	}

	// Converter theme (usar default se não fornecido)
	theme := room.Theme(req.Theme)
	if theme == "" {
		theme = room.ThemeDefault
	}

	// Converter visibility (usar public se não fornecido)
	visibility := room.Visibility(req.Visibility)
	if visibility == "" {
		visibility = room.VisibilityPublic
	}

	// Criar a sala
	output, err := h.roomService.Create(r.Context(), approom.CreateInput{
		OwnerID:    user.ID(userID),
		Name:       req.Name,
		Theme:      theme,
		Visibility: visibility,
	})

	if err != nil {
		handleRoomError(w, err)
		return
	}

	// Retornar a sala criada (incluindo o código para o dono)
	httputil.JSON(w, http.StatusCreated, toRoomResponse(output.Room, true))
}

// ListPublic lista as salas públicas.
// GET /api/v1/rooms
func (h *RoomHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	// TODO: Pegar limit e offset da query string
	rooms, err := h.roomService.ListPublic(r.Context(), approom.ListPublicInput{
		Limit:  20,
		Offset: 0,
	})

	if err != nil {
		httputil.InternalServerError(w, "Failed to list rooms")
		return
	}

	// Converter para response (sem código de acesso)
	response := make([]RoomResponse, len(rooms))
	for i, rm := range rooms {
		response[i] = toRoomResponse(rm, false)
	}

	httputil.JSON(w, http.StatusOK, response)
}

// ListMy lista as salas do usuário autenticado.
// GET /api/v1/rooms/my
func (h *RoomHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserID(r.Context())
	if userID == "" {
		httputil.Unauthorized(w, "User not authenticated")
		return
	}

	rooms, err := h.roomService.ListByOwner(r.Context(), user.ID(userID))
	if err != nil {
		httputil.InternalServerError(w, "Failed to list rooms")
		return
	}

	// Converter para response (com código, pois é o dono)
	response := make([]RoomResponse, len(rooms))
	for i, rm := range rooms {
		response[i] = toRoomResponse(rm, true)
	}

	httputil.JSON(w, http.StatusOK, response)
}

// GetByID busca uma sala pelo ID.
// GET /api/v1/rooms/:id
func (h *RoomHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	if roomID == "" {
		httputil.BadRequest(w, "Room ID is required")
		return
	}

	rm, err := h.roomService.GetByID(r.Context(), room.ID(roomID))
	if err != nil {
		handleRoomError(w, err)
		return
	}

	// Verificar se o usuário é o dono para mostrar o código
	userID := httputil.GetUserID(r.Context())
	isOwner := userID != "" && rm.IsOwner(user.ID(userID))

	httputil.JSON(w, http.StatusOK, toRoomResponse(rm, isOwner))
}

// JoinByCodeRequest é o corpo da requisição de entrar por código.
type JoinByCodeRequest struct {
	AccessCode string `json:"access_code"`
}

// JoinByCode busca uma sala pelo código de acesso.
// POST /api/v1/rooms/join
func (h *RoomHandler) JoinByCode(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserID(r.Context())
	if userID == "" {
		httputil.Unauthorized(w, "User not authenticated")
		return
	}

	var req JoinByCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Invalid request body")
		return
	}

	if req.AccessCode == "" {
		httputil.BadRequest(w, "Access code is required")
		return
	}

	rm, err := h.roomService.JoinByCode(r.Context(), approom.JoinByCodeInput{
		AccessCode: req.AccessCode,
		UserID:     user.ID(userID),
	})

	if err != nil {
		handleRoomError(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, toRoomResponse(rm, false))
}

// Delete deleta uma sala.
// DELETE /api/v1/rooms/:id
func (h *RoomHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := httputil.GetUserID(r.Context())
	if userID == "" {
		httputil.Unauthorized(w, "User not authenticated")
		return
	}

	roomID := chi.URLParam(r, "id")
	if roomID == "" {
		httputil.BadRequest(w, "Room ID is required")
		return
	}

	err := h.roomService.Delete(r.Context(), approom.DeleteInput{
		RoomID:      room.ID(roomID),
		RequesterID: user.ID(userID),
	})

	if err != nil {
		handleRoomError(w, err)
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"message": "Room deleted successfully"})
}

// handleRoomError trata erros do serviço de room.
func handleRoomError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, approom.ErrMaxRoomsReached):
		httputil.BadRequest(w, "You can only have 2 rooms at a time")
	case errors.Is(err, approom.ErrRoomNotFound):
		httputil.NotFound(w, "Room not found")
	case errors.Is(err, approom.ErrNotRoomOwner):
		httputil.Forbidden(w, "You are not the owner of this room")
	case errors.Is(err, approom.ErrInvalidCode):
		httputil.NotFound(w, "Invalid access code")
	case errors.Is(err, room.ErrNameTooShort):
		httputil.BadRequest(w, "Room name must be at least 3 characters")
	case errors.Is(err, room.ErrNameTooLong):
		httputil.BadRequest(w, "Room name must be at most 25 characters")
	case errors.Is(err, room.ErrInvalidTheme):
		httputil.BadRequest(w, "Invalid theme")
	case errors.Is(err, room.ErrInvalidVisibility):
		httputil.BadRequest(w, "Invalid visibility (use 'public' or 'private')")
	case errors.Is(err, room.ErrRoomNotEmpty):
		httputil.BadRequest(w, "Room must be empty to delete")
	default:
		httputil.InternalServerError(w, "An unexpected error occurred")
	}
}
