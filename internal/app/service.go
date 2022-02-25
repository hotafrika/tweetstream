package app

import (
	"github.com/hotafrika/tweetstream/internal/domain"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
	"github.com/rs/zerolog"
	"reflect"
	"sync"
)

// Service works as a broker between the source of tweets and tweets consumers
type Service struct {
	sourcer   domain.Sourcer
	consumers []domain.Consumer
	logger    *zerolog.Logger
}

// NewService creates new Service with initialized Sourcer and consumers
func NewService(s domain.Sourcer, logger *zerolog.Logger, c ...domain.Consumer) Service {
	service := Service{
		sourcer:   s,
		consumers: c,
		logger:    logger,
	}
	return service
}

// Run starts service to act as tweets broker.
// It also starts Sourcer to produce tweets and consumers to consume tweets.
func (s Service) Run(quitCh chan struct{}) error {
	var wg sync.WaitGroup
	// start Sourcer
	sourceQuit := make(chan struct{})
	sourceCh, err := s.sourcer.Get(sourceQuit)
	if err != nil {
		return err
	}

	// start consumers
	chArr := make([]chan entities.Tweet, len(s.consumers))
	for i, consumer := range s.consumers {
		// size 1 for faster distribution
		destCh := make(chan entities.Tweet, 1)
		chArr[i] = destCh
		wg.Add(1)
		consumer := consumer
		go func() {
			consumer.Run(destCh)
			wg.Done()
		}()
	}

Loop:
	for {
		select {
		case tweet, ok := <-sourceCh:
			// if sourceCh was closed
			if !ok {
				s.logger.Info().Str("Unit", reflect.TypeOf(s).String()).Msg("source channel was closed")
				for _, ch := range chArr {
					close(ch)
				}
				break Loop
			}
			// this is the bottleneck
			// because the common speed of new tweets processing depends on the slowest consumer
			for _, ch := range chArr {
				ch <- tweet
			}
		case <-quitCh:
			s.logger.Info().Str("Unit", reflect.TypeOf(s).String()).Msg("got a quit signal via quit channel")
			close(sourceQuit)
			for _, ch := range chArr {
				close(ch)
			}
			break Loop
		}
	}

	// wait until all consumers stop tweets processing
	wg.Wait()
	return nil
}
