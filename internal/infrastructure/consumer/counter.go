package consumer

import (
	"github.com/hotafrika/tweetstream/internal/domain"
	"github.com/hotafrika/tweetstream/internal/domain/entities"
)

var _ domain.Consumer = (*Counter)(nil)

// Counter is a simple consumer which counts new tweets
type Counter struct {
	count int
}

// Run ...
func (c *Counter) Run(inputCh chan entities.Tweet) {
	for _ = range inputCh {
		c.count++
	}
}

// GetCount returns count of processed tweets
func (c *Counter) GetCount() int {
	return c.count
}
