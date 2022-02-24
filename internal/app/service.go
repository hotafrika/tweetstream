package app

import (
	"github.com/hotafrika/tweetstream/internal/domain"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
	"sync"
)

// Service works as broker between the source of tweets and tweets consumer
type Service struct {
	Source    domain.Sourcer
	Consumers []domain.Consumer
}

// NewService creates new Service with initialized Sourcer and Consumers
func NewService(s domain.Sourcer, c ...domain.Consumer) Service {
	service := Service{
		Source:    s,
		Consumers: c,
	}
	return service
}

// Start starts service act as tweets broker.
// It also starts Sourcer to produce tweets and Consumers to consume tweets.
func (s Service) Start(quitCh chan struct{}) error {
	var wg sync.WaitGroup
	sourceQuit := make(chan struct{})
	sourceCh, err := s.Source.Get(sourceQuit)
	if err != nil {
		return err
	}
	chArr := make([]chan entities.Tweet, len(s.Consumers))
	for i, consumer := range s.Consumers {
		destCh := make(chan entities.Tweet, 1)
		chArr[i] = destCh
		wg.Add(1)
		consumer := consumer
		go func() {
			consumer.Start(destCh)
			wg.Done()
		}()
	}

Loop:
	for {
		select {
		case tweet, ok := <-sourceCh:
			if !ok {
				for _, ch := range chArr {
					close(ch)
				}
				break Loop
			}
			for _, ch := range chArr {
				// TODO maybe go
				ch <- tweet
			}
		case <-quitCh:
			close(sourceQuit)
			for _, ch := range chArr {
				close(ch)
			}
			break Loop
		}
	}

	// wait until all consumer stop tweets processing
	wg.Wait()
	return nil
}
