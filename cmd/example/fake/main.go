package main

import (
	"github.com/hotafrika/tweetstream/internal/app"
	"github.com/hotafrika/tweetstream/internal/infrastructure/consumer"
	"github.com/hotafrika/tweetstream/internal/infrastructure/sourcer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

func main() {
	logger := log.Level(zerolog.DebugLevel)
	s := sourcer.NewFake(2*time.Second, "tweets.json", &logger)
	c1 := consumer.Console{}

	service := app.NewService(s, &logger, c1)
	quit := make(chan struct{})
	err := service.Run(quit)
	if err != nil {
		panic(err)
	}
}
