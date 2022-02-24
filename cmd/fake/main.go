package main

import (
	"github.com/hotafrika/tweetstream/internal/app"
	"github.com/hotafrika/tweetstream/internal/infrastructure/consumer"
	"github.com/hotafrika/tweetstream/internal/infrastructure/sourcer"
	"time"
)

func main() {
	s := sourcer.NewFake(2*time.Second, "tweets.json")
	c1 := consumer.Console{}

	service := app.NewService(s, c1)
	quit := make(chan struct{})
	err := service.Start(quit)
	if err != nil {
		panic(err)
	}
}
