package consumer

import (
	"fmt"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
)

// Console is a simple consumer which prints new tweets to console output
type Console struct {
}

// Run ...
func (c Console) Run(inputCh chan entities.Tweet) {
	for tweet := range inputCh {
		fmt.Println(tweet)
	}
}
