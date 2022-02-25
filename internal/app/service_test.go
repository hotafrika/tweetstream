package app

import (
	"github.com/hotafrika/tweetstream/internal/infrastructure/consumer"
	"github.com/hotafrika/tweetstream/internal/infrastructure/sourcer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
	"time"
)

func TestService_RunLimited(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		sleep     time.Duration
		timeout   time.Duration
		wantCount int
	}{
		{
			name:      "limited",
			filename:  "limited.json",
			sleep:     1 * time.Millisecond,
			timeout:   1 * time.Second,
			wantCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := log.Level(zerolog.ErrorLevel)
			s := sourcer.NewFakeL(tt.sleep, path.Join("testdata", tt.filename), &l)
			c := consumer.Counter{}
			service := NewService(s, &l, &c)

			timer := time.NewTimer(tt.timeout)
			quit := make(chan struct{})
			errChan := make(chan error)
			go func() {
				errChan <- service.Run(quit)
			}()

			select {
			// if return from service
			case err := <-errChan:
				if !assert.NoError(t, err) {
					return
				}
				assert.Equal(t, tt.wantCount, c.GetCount())
			// if return from timer
			case <-timer.C:
				quit <- struct{}{}
				assert.Fail(t, "service timeout")
			}
		})
	}
}

func TestService_RunUnlimited(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		sleep    time.Duration
		timeout  time.Duration
	}{
		{
			name:     "unlimited",
			filename: "limited.json",
			sleep:    1 * time.Millisecond,
			timeout:  100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := log.Level(zerolog.ErrorLevel)
			s := sourcer.NewFake(tt.sleep, path.Join("testdata", tt.filename), &l)
			c := consumer.Counter{}
			service := NewService(s, &l, &c)

			timer := time.NewTimer(2 * tt.timeout)
			quit := make(chan struct{})
			errChan := make(chan error)
			go func() {
				errChan <- service.Run(quit)
			}()

			go func() {
				time.Sleep(tt.timeout)
				close(quit)
			}()

			select {
			case err := <-errChan:
				if !assert.NoError(t, err) {
					return
				}
				assert.NotEqual(t, 0, c.GetCount())
			case <-timer.C:
				assert.Fail(t, "service timeout")
			}
		})
	}
}
