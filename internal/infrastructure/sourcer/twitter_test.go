package sourcer

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTwitter_Get_WrongToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "wrong token",
			token: "abc",
		},
	}
	for _, tt := range tests {
		logger := log.Level(zerolog.ErrorLevel)
		t.Run(tt.name, func(t *testing.T) {
			twitter := NewTwitter(tt.token, &logger)
			quit := make(chan struct{})
			ch, err := twitter.Get(quit)
			if !assert.Error(t, err) {
				close(ch)
			}
		})
	}
}
