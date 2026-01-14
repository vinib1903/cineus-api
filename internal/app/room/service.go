package room

import (
	"context"
	"errors"

	"github.com/vinib1903/cineus-api/internal/domain/room"
	"github.com/vinib1903/cineus-api/internal/domain/user"
	"github.com/vinib1903/cineus-api/internal/infra/auth"
)

// Erros do serviço de room.
var (
	ErrMaxRoomsReached = errors.New("maximum number of rooms reached (2)")
	ErrRoomNotFound    = errors.New("room not found")
	ErrNotRoomOwner    = errors.New("you are not the owner of this room")
	ErrInvalidCode     = errors.New("invalid access code")
)

// MaxRoomsPerUser é o limite de salas por usuário.
const MaxRoomsPerUser = 2

// Service contém a lógica de negócio de salas.
type Service struct {
	roomRepo room.Repository
	idGen    *auth.IDGenerator
}

// NewService cria uma nova instância do serviço.
func NewService(roomRepo room.Repository, idGen *auth.IDGenerator) *Service {
	return &Service{
		roomRepo: roomRepo,
		idGen:    idGen,
	}
}

// CreateInput são os dados para criar uma sala.
type CreateInput struct {
	OwnerID    user.ID
	Name       string
	Theme      room.Theme
	Visibility room.Visibility
}

// CreateOutput é o resultado da criação.
type CreateOutput struct {
	Room *room.Room
}

// Create cria uma nova sala.
func (s *Service) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	// Verificar limite de salas por usuário
	count, err := s.roomRepo.CountByOwner(ctx, input.OwnerID)
	if err != nil {
		return nil, err
	}
	if count >= MaxRoomsPerUser {
		return nil, ErrMaxRoomsReached
	}

	// Gerar ID
	roomID := room.ID(s.idGen.NewID())

	// Criar a sala
	newRoom, err := room.NewRoom(roomID, input.OwnerID, input.Name, input.Theme, input.Visibility)
	if err != nil {
		return nil, err
	}

	// Salvar no banco
	if err := s.roomRepo.Create(ctx, newRoom); err != nil {
		return nil, err
	}

	return &CreateOutput{Room: newRoom}, nil
}

// ListPublicInput são os dados para listar salas públicas.
type ListPublicInput struct {
	Limit  int
	Offset int
}

// ListPublic retorna as salas públicas.
func (s *Service) ListPublic(ctx context.Context, input ListPublicInput) ([]*room.Room, error) {
	// Valores padrão
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	return s.roomRepo.ListPublic(ctx, input.Limit, input.Offset)
}

// ListByOwner retorna as salas de um usuário.
func (s *Service) ListByOwner(ctx context.Context, ownerID user.ID) ([]*room.Room, error) {
	return s.roomRepo.ListByOwner(ctx, ownerID)
}

// GetByID busca uma sala pelo ID.
func (s *Service) GetByID(ctx context.Context, id room.ID) (*room.Room, error) {
	r, err := s.roomRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, room.ErrRoomNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}
	return r, nil
}

// JoinByCodeInput são os dados para entrar em uma sala por código.
type JoinByCodeInput struct {
	AccessCode string
	UserID     user.ID
}

// JoinByCode busca uma sala pelo código de acesso.
func (s *Service) JoinByCode(ctx context.Context, input JoinByCodeInput) (*room.Room, error) {
	r, err := s.roomRepo.GetByAccessCode(ctx, input.AccessCode)
	if err != nil {
		if errors.Is(err, room.ErrRoomNotFound) {
			return nil, ErrInvalidCode
		}
		return nil, err
	}

	// TODO: Verificar se usuário está banido
	// TODO: Verificar se sala está cheia

	return r, nil
}

// DeleteInput são os dados para deletar uma sala.
type DeleteInput struct {
	RoomID      room.ID
	RequesterID user.ID
}

// Delete deleta uma sala.
func (s *Service) Delete(ctx context.Context, input DeleteInput) error {
	// Buscar a sala
	r, err := s.roomRepo.GetByID(ctx, input.RoomID)
	if err != nil {
		if errors.Is(err, room.ErrRoomNotFound) {
			return ErrRoomNotFound
		}
		return err
	}

	// Verificar se é o dono
	if !r.IsOwner(input.RequesterID) {
		return ErrNotRoomOwner
	}

	// TODO: Verificar se a sala está vazia
	// Por enquanto, assumimos que está vazia
	isEmpty := true

	// Deletar (soft delete)
	if err := r.Delete(input.RequesterID, isEmpty); err != nil {
		return err
	}

	// Atualizar no banco
	return s.roomRepo.Update(ctx, r)
}
