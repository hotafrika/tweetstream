package sourcer

import (
	"encoding/json"
	"errors"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
	"io"
	"os"
	"time"
)

type Fake struct {
	sleep    time.Duration
	filename string
}

func NewFake(sleep time.Duration, filename string) Fake {
	return Fake{
		sleep:    sleep,
		filename: filename,
	}
}

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
		i := 0
	Loop:
		for {
			i = i % len(tweets)
			select {
			case <-quit:
				break Loop
			case destCh <- tweets[i]:
				i++
				time.Sleep(f.sleep)
			}
		}
	}()

	return destCh, nil
}
