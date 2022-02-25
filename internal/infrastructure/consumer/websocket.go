package consumer

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
	"github.com/rs/zerolog"
	"net/http"
	"reflect"
	"time"
)

// WS represents websocket consumer
type WS struct {
	server       *http.Server
	logger       *zerolog.Logger
	clients      map[*websocket.Conn]struct{}
	registerCh   chan *websocket.Conn
	deregisterCh chan *websocket.Conn
}

// NewWS creates new WS. Also it initializes server and necessary channels
func NewWS(addr string, logger *zerolog.Logger) *WS {
	clients := make(map[*websocket.Conn]struct{})
	registerCh := make(chan *websocket.Conn)
	deregisterCh := make(chan *websocket.Conn)
	server := &http.Server{Addr: addr}
	ws := &WS{
		server:       server,
		logger:       logger,
		clients:      clients,
		registerCh:   registerCh,
		deregisterCh: deregisterCh,
	}

	ws.registerRoutes()

	return ws
}

func (ws *WS) registerRoutes() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.handleWS)
	ws.server.Handler = mux
}

var upgrader = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ws *WS) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Warn().Str("Unit", reflect.TypeOf(ws).String()).Err(err).Msg("failed to upgrade websocket")
		return
	}
	ws.Add(conn)
	defer ws.Delete(conn)
	// infinitely listen messages. do not process them. handle disconnect
	for {
		mt, _, err := conn.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			break
		}
	}
}

// Add adds the connection to the clients list
func (ws *WS) Add(c *websocket.Conn) {
	ws.registerCh <- c
}

// Delete removes the connection from the clients list
func (ws *WS) Delete(c *websocket.Conn) {
	ws.deregisterCh <- c
}

// processEvents acts as event processor: quit event, add connection, delete connection and new tweet dispatch
func (ws *WS) processEvents(quitCh chan struct{}, dispatchCh chan entities.Tweet) {
	for {
		select {
		// close all connections and stop processing
		case <-quitCh:
			ws.logger.Info().Str("Unit", reflect.TypeOf(ws).String()).Msg("closing all connections")
			for conn := range ws.clients {
				err := conn.Close()
				if err != nil {
					ws.logger.Warn().Str("Unit", reflect.TypeOf(ws).String()).Err(err).Msg("failed to close connection")
				}
			}
			return
		// dispatch tweet to all clients
		case tweet := <-dispatchCh:
			ws.logger.Info().Str("Unit", reflect.TypeOf(ws).String()).Msg("new tweet dispatching")
			for conn := range ws.clients {
				err := conn.WriteJSON(tweet)
				if err != nil {
					ws.logger.Warn().Str("Unit", reflect.TypeOf(ws).String()).Err(err).Msg("failed to write JSON")
				}
			}
		case conn := <-ws.registerCh:
			ws.logger.Info().Str("Unit", reflect.TypeOf(ws).String()).Msg("new client connected")
			ws.clients[conn] = struct{}{}
		case conn := <-ws.deregisterCh:
			ws.logger.Info().Str("Unit", reflect.TypeOf(ws).String()).Msg("client disconnected")
			delete(ws.clients, conn)
		}
	}
}

// Run starts reading of new tweets from source channel.
func (ws *WS) Run(sourceCh chan entities.Tweet) {
	quitCh := make(chan struct{})
	dispatchCh := make(chan entities.Tweet)

	// start http server
	go func() {
		ws.logger.Info().Str("Unit", reflect.TypeOf(ws).String()).Msg("server started")
		err := ws.server.ListenAndServe()
		ws.logger.Warn().Str("Unit", reflect.TypeOf(ws).String()).Err(err).Msg("server stopped")
	}()

	// start event processing
	go ws.processEvents(quitCh, dispatchCh)

	// listen source channel until it is closed
	for tweet := range sourceCh {
		dispatchCh <- tweet
	}

	// when source channel is closed
	// stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := ws.server.Shutdown(ctx)
	if err != nil {
		ws.logger.Warn().Str("Unit", reflect.TypeOf(ws).String()).Err(err).Msg("server shutdown error")
	}
	// and close all connections
	close(quitCh)
}
