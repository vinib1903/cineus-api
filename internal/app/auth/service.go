package auth

import (
	"context"
	"errors"

	"github.com/vinib1903/cineus-api/internal/domain/user"
	"github.com/vinib1903/cineus-api/internal/infra/auth"
)

// Erros do serviço de autenticação.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already registered")
)

// Service contém a lógica de negócio de autenticação.
type Service struct {
	userRepo user.Repository
	hasher   *auth.PasswordHasher
	jwt      *auth.JWTManager
	idGen    *auth.IDGenerator
}

// NewService cria uma nova instância do serviço.
func NewService(
	userRepo user.Repository,
	hasher *auth.PasswordHasher,
	jwt *auth.JWTManager,
	idGen *auth.IDGenerator,
) *Service {
	return &Service{
		userRepo: userRepo,
		hasher:   hasher,
		jwt:      jwt,
		idGen:    idGen,
	}
}

// RegisterInput são os dados necessários para registro.
type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

// RegisterOutput é o resultado do registro.
type RegisterOutput struct {
	User   *user.User
	Tokens *auth.TokenPair
}

// Register cria uma nova conta de usuário.
func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	// Validar senha
	if err := user.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	// Verificar se email já existe
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Gerar hash da senha
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	// Gerar ID único
	userID := user.ID(s.idGen.NewID())

	// Criar entidade de usuário
	newUser, err := user.NewUser(userID, input.Email, passwordHash, input.DisplayName)
	if err != nil {
		return nil, err
	}

	// Salvar no banco
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	// Gerar tokens
	tokens, err := s.jwt.GenerateTokenPair(string(newUser.ID), newUser.Email)
	if err != nil {
		return nil, err
	}

	return &RegisterOutput{
		User:   newUser,
		Tokens: tokens,
	}, nil
}

// LoginInput são os dados necessários para login.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput é o resultado do login.
type LoginOutput struct {
	User   *user.User
	Tokens *auth.TokenPair
}

// Login autentica um usuário existente.
func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// Buscar usuário pelo email
	existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verificar senha
	if err := s.hasher.Compare(existingUser.PasswordHash, input.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Registrar o login
	existingUser.RecordLogin()
	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, err
	}

	// Gerar tokens
	tokens, err := s.jwt.GenerateTokenPair(string(existingUser.ID), existingUser.Email)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		User:   existingUser,
		Tokens: tokens,
	}, nil
}
