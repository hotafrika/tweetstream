package consumer

import (
	"fmt"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
)

type Console struct {
}

func (c Console) Start(inputCh chan entities.Tweet) {
	for tweet := range inputCh {
		fmt.Println(tweet)
	}
}
