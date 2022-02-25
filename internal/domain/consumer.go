package domain

import "github.com/hotafrika/tweetstream/internal/domain/entities"

// Consumer is an interface for any kind of tweet consumer
// Init is a method for consumer initialization with input channel. Consumer must process new tweets almost in real time.
// Or it must create own buffer.
type Consumer interface {
	Run(chan entities.Tweet)
}
