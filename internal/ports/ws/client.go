package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	// Tempo máximo para escrever uma mensagem
	writeWait = 10 * time.Second

	// Tempo máximo para ler uma mensagem (pong)
	pongWait = 60 * time.Second

	// Intervalo para enviar pings
	pingPeriod = (pongWait * 9) / 10

	// Tamanho máximo da mensagem
	maxMessageSize = 4096

	// Tamanho do buffer do canal de envio
	sendBufferSize = 256
)

// Client representa uma conexão WebSocket de um usuário.
type Client struct {
	// Hub da sala em que o cliente está
	hub *RoomHub

	// Conexão WebSocket
	conn *websocket.Conn

	// Canal para enviar mensagens (buffered)
	send chan []byte

	// Informações do usuário
	userID      string
	displayName string
	seatID      string

	// Mutex para proteger o seatID
	mu sync.RWMutex

	// Contexto para cancelamento
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient cria um novo cliente.
func NewClient(hub *RoomHub, conn *websocket.Conn, userID, displayName string) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, sendBufferSize),
		userID:      userID,
		displayName: displayName,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// GetUserID retorna o ID do usuário.
func (c *Client) GetUserID() string {
	return c.userID
}

// GetDisplayName retorna o nome de exibição.
func (c *Client) GetDisplayName() string {
	return c.displayName
}

// GetSeatID retorna o ID do assento (thread-safe).
func (c *Client) GetSeatID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.seatID
}

// SetSeatID define o ID do assento (thread-safe).
func (c *Client) SetSeatID(seatID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seatID = seatID
}

// Run inicia as goroutines de leitura e escrita.
func (c *Client) Run() {
	// Inicia a goroutine de escrita
	go c.writePump()

	// Executa a leitura na goroutine atual
	c.readPump()
}

// readPump lê mensagens do WebSocket e processa.
func (c *Client) readPump() {
	// Quando sair desta função, limpa tudo
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "connection closed")
		c.cancel()
	}()

	// Configurar limite de tamanho
	c.conn.SetReadLimit(maxMessageSize)

	for {
		// Ler mensagem
		msgType, data, err := c.conn.Read(c.ctx)
		if err != nil {
			// Conexão fechada ou erro
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Printf("Client %s disconnected normally", c.userID)
			} else {
				log.Printf("Client %s read error: %v", c.userID, err)
			}
			return
		}

		// Só processamos mensagens de texto (JSON)
		if msgType != websocket.MessageText {
			continue
		}

		// Parsear a mensagem
		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Client %s: invalid message format: %v", c.userID, err)
			c.SendError("INVALID_FORMAT", "Invalid message format")
			continue
		}

		// Processar a mensagem
		c.hub.handleMessage(c, &msg)
	}
}

// writePump envia mensagens do canal para o WebSocket.
func (c *Client) writePump() {
	// Ticker para enviar pings
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "write pump closed")
	}()

	for {
		select {
		case <-c.ctx.Done():
			// Contexto cancelado, sair
			return

		case message, ok := <-c.send:
			if !ok {
				// Canal fechado, sair
				return
			}

			// Definir deadline para escrita
			ctx, cancel := context.WithTimeout(c.ctx, writeWait)

			// Enviar a mensagem
			err := c.conn.Write(ctx, websocket.MessageText, message)
			cancel()

			if err != nil {
				log.Printf("Client %s write error: %v", c.userID, err)
				return
			}

		case <-ticker.C:
			// Enviar ping para manter a conexão viva
			ctx, cancel := context.WithTimeout(c.ctx, writeWait)
			err := c.conn.Ping(ctx)
			cancel()

			if err != nil {
				log.Printf("Client %s ping error: %v", c.userID, err)
				return
			}
		}
	}
}

// Send envia uma mensagem para o cliente.
func (c *Client) Send(msg *OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Client %s: failed to marshal message: %v", c.userID, err)
		return
	}

	// Tenta enviar, mas não bloqueia se o buffer estiver cheio
	select {
	case c.send <- data:
		// Enviado com sucesso
	default:
		// Buffer cheio - cliente está muito lento
		log.Printf("Client %s: send buffer full, closing connection", c.userID)
		c.cancel()
	}
}

// SendError envia uma mensagem de erro.
func (c *Client) SendError(code, message string) {
	c.Send(NewOutgoingMessage(TypeError, ErrorPayload{
		Code:    code,
		Message: message,
	}))
}

// Close fecha a conexão do cliente.
func (c *Client) Close() {
	c.cancel()
	close(c.send)
}
