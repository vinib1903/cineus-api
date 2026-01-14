package http

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/vinib1903/cineus-api/internal/ports/http/httputil"
)

// Logger é um middleware que loga as requisições.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Cria um wrapper para capturar o status code
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Chama o próximo handler
		next.ServeHTTP(wrapped, r)

		// Loga a requisição
		log.Printf(
			"%s %s %d %s",
			r.Method,
			r.URL.Path,
			wrapped.status,
			time.Since(start),
		)
	})
}

// responseWriter é um wrapper que captura o status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader captura o status code antes de escrever.
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack implementa http.Hijacker para suportar WebSocket.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// Verifica se o ResponseWriter original suporta Hijack
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

// Flush implementa http.Flusher para suportar streaming.
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Recoverer é um middleware que recupera de panics.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				httputil.InternalServerError(w, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORS adiciona headers para permitir requisições cross-origin.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Não aplicar CORS em WebSocket upgrades
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		// Permite qualquer origem em desenvolvimento
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Responde imediatamente para requisições OPTIONS (preflight)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
