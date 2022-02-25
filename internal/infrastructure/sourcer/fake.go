package sourcer

import (
	"encoding/json"
	"errors"
	"github.com/hotafrika/tweetstream/internal/domain"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
	"github.com/rs/zerolog"
	"io"
	"os"
	"reflect"
	"time"
)

// Fake is a simple Sourcer which reads tweets from file and infinitely send it
type Fake struct {
	sleep    time.Duration
	filename string
	logger   *zerolog.Logger
}

var _ domain.Sourcer = Fake{}

// NewFake creates new Fake
func NewFake(sleep time.Duration, filename string, logger *zerolog.Logger) Fake {
	return Fake{
		sleep:    sleep,
		filename: filename,
		logger:   logger,
	}
}

// Get starts Fake to send tweets from the file
func (f Fake) Get(quit chan struct{}) (chan entities.Tweet, error) {
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
		i := 0
		for {
			i = i % len(tweets)
			select {
			case <-quit:
				f.logger.Info().Str("Unit", reflect.TypeOf(f).String()).Msg("got quit signal via quit channel")
				return
			case destCh <- tweets[i]:
				i++
				time.Sleep(f.sleep)
			}
		}
	}()

	return destCh, nil
}
