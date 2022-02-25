package consumer

import (
	"github.com/gorilla/websocket"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestWS_Run(t *testing.T) {
	returnCh := make(chan int)

	// create WS
	logger := log.Level(zerolog.DebugLevel)
	ws := NewWS("", &logger)

	// create custom server with listener
	server := httptest.NewUnstartedServer(http.HandlerFunc(ws.handleWS))
	server.Start()

	// create source channel
	sourceCh := make(chan entities.Tweet)
	go ws.Run(sourceCh)

	// get URL of custom server
	uTemp, err := url.Parse(server.URL)
	if err != nil {
		server.Close()
		close(sourceCh)
		assert.NoError(t, err)
		return
	}

	// create websocket client
	u := url.URL{Scheme: "ws", Host: uTemp.Host, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		server.Close()
		close(sourceCh)
		assert.NoError(t, err)
		return
	}

	// read messages from websocket and count them
	go func() {
		count := 0
		defer func() {
			returnCh <- count
		}()
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				return
			}
			count++
		}
	}()

	// send 2 tweets
	sourceCh <- entities.Tweet{}
	sourceCh <- entities.Tweet{}

	server.Close()
	close(sourceCh)

	// get tweets counts from client side
	count := <-returnCh

	// expect 2 tweets
	assert.Equal(t, 2, count)
}
