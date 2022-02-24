package domain

import "github.com/hotafrika/tweetstream/internal/domain/entities"

// Sourcer is an iterface for tweets source
// Get returns channer for getting tweets and error
type Sourcer interface {
	Get(chan struct{}) (chan entities.Tweet, error)
}
