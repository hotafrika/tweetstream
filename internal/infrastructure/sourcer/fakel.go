package sourcer

import (
	"encoding/json"
	"errors"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
	"github.com/rs/zerolog"
	"io"
	"os"
	"reflect"
	"time"
)

// FakeL is a simple Sourcer which reads tweets from file and send it
// It stops execution when all the tweets are sent
type FakeL struct {
	sleep    time.Duration
	filename string
	logger   *zerolog.Logger
}

// NewFakeL creates new FakeL
func NewFakeL(sleep time.Duration, filename string, logger *zerolog.Logger) FakeL {
	return FakeL{
		sleep:    sleep,
		filename: filename,
		logger:   logger,
	}
}

// Get starts FakeL to send tweets from the file
// Returns when all the tweets are sent
func (f FakeL) Get(quit chan struct{}) (chan entities.Tweet, error) {
	file, err := os.Open(f.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var tweets []entities.Tweet
	err = json.Unmarshal(b, &tweets)
	if err != nil {
		return nil, err
	}

	if len(tweets) == 0 {
		return nil, errors.New("tweets len is 0")
	}

	destCh := make(chan entities.Tweet)
	go func() {
		f.logger.Info().Str("Unit", reflect.TypeOf(f).String()).Msg("start sending tweets")
		defer func() {
			f.logger.Info().Str("Unit", reflect.TypeOf(f).String()).Msg("close tweets source channel")
			close(destCh)
		}()
		for _, tweet := range tweets {
			select {
			case <-quit:
				f.logger.Info().Str("Unit", reflect.TypeOf(f).String()).Msg("got quit signal via quit channel")
				return
			case destCh <- tweet:
				time.Sleep(f.sleep)
			}
		}
	}()

	return destCh, nil
}
