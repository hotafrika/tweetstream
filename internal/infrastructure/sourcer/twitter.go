package sourcer

import (
	"encoding/json"
	"github.com/pkg/errors"
	"reflect"
	"time"

	twitterstream "github.com/fallenstedt/twitter-stream"
	"github.com/hotafrika/tweetstream/internal/domain"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
	"github.com/rs/zerolog"
)

// TweetData is used for parsing raw tweet messages
type TweetData struct {
	Data struct {
		Text      string    `json:"text"`
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		AuthorID  string    `json:"author_id"`
	} `json:"data"`
	Includes struct {
		Users []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Username string `json:"username"`
			URL      string `json:"url"`
			Location string `json:"location"`
		} `json:"users"`
	} `json:"includes"`
}

// ToTweet casts TweetData to entities.Tweet object
func (td TweetData) ToTweet() entities.Tweet {
	author := entities.Author{}
	for _, u := range td.Includes.Users {
		if td.Data.AuthorID == u.ID {
			author.ID = u.ID
			author.Name = u.Name
			author.Username = u.Username
			author.URL = u.URL
			author.Location = u.Location
		}
	}

	return entities.Tweet{
		Text:      td.Data.Text,
		ID:        td.Data.ID,
		CreatedAt: td.Data.CreatedAt,
		Author:    author,
	}
}

// Twitter is an implementation of Twitter Filtered Stream source of tweets.
// It uses long polling for getting new messages.
// To use this Sourcer you need at least one rule configured.
type Twitter struct {
	api    *twitterstream.TwitterApi
	logger *zerolog.Logger
}

var _ domain.Sourcer = (*Twitter)(nil)

// NewTwitter creates Twitter Sourcer
func NewTwitter(token string, logger *zerolog.Logger) *Twitter {
	api := twitterstream.NewTwitterStream(token)

	api.Stream.SetUnmarshalHook(func(bytes []byte) (interface{}, error) {
		logger.Debug().Str("debug object UnmarshalHook", string(bytes)).Msg("")
		data := TweetData{}
		err := json.Unmarshal(bytes, &data)
		if err != nil {
			logger.Warn().Err(err).Str("debug object", string(bytes)).Msg("failed to unmarshal bytes in UnmarshalHook")
		}
		return data, err
	})

	return &Twitter{
		api:    api,
		logger: logger,
	}
}

// Get starts getting of new tweets from Twitter API
func (t *Twitter) Get(quitCh chan struct{}) (chan entities.Tweet, error) {
	destCh := make(chan entities.Tweet)

	streamExpansions := twitterstream.NewStreamQueryParamsBuilder().
		AddTweetField("created_at").
		AddExpansion("author_id").
		AddExpansion("geo.place_id").
		AddUserField("location").
		AddUserField("name").
		AddUserField("url").
		Build()

	// Run the Stream
	t.logger.Info().Str("Unit", reflect.TypeOf(t).String()).Msg("start sending tweets")
	err := t.api.Stream.StartStream(streamExpansions)
	if err != nil {
		close(destCh)
		return nil, errors.Wrap(err, "StartStream")
	}

	go func() {
		defer func() {
			close(destCh)
			t.logger.Info().Str("Unit", reflect.TypeOf(t).String()).Msg("output channel was closed")
		}()

		for {
			select {
			case <-quitCh:
				t.logger.Info().Str("Unit", reflect.TypeOf(t).String()).Msg("got quit signal")
				t.api.Stream.StopStream()
				return
			case message, ok := <-t.api.Stream.GetMessages():
				if !ok {
					t.logger.Info().Str("Unit", reflect.TypeOf(t).String()).Msg("source channel was closed")
					return
				}
				if message.Err != nil {
					t.logger.Warn().Str("Unit", reflect.TypeOf(t).String()).Err(message.Err).Msg("got error from twitter")
					continue
				}
				// here we need to check again if quit signal could come
				select {
				case <-quitCh:
					t.logger.Info().Str("Unit", reflect.TypeOf(t).String()).Msg("got quit signal")
					t.api.Stream.StopStream()
					return
				// don't need to check if type assertion was successful because we use TweetData in SetUnmarshalHook()
				case destCh <- message.Data.(TweetData).ToTweet():
				}
			}
		}
	}()

	return destCh, nil
}
