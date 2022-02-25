package main

import (
	"github.com/hotafrika/tweetstream/internal/app"
	"github.com/hotafrika/tweetstream/internal/infrastructure/consumer"
	"github.com/hotafrika/tweetstream/internal/infrastructure/sourcer"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var myEnv map[string]string
	myEnv, err := godotenv.Read()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	wsAddr := myEnv["WS_ADDRESS"]
	twitterToken := myEnv["TWITTER_TOKEN"]

	logger := zlog.Level(zerolog.DebugLevel)

	// consumers
	ws := consumer.NewWS(wsAddr, &logger)
	console := consumer.Console{}
	// source
	twitter := sourcer.NewTwitter(twitterToken, &logger)
	// service
	service := app.NewService(twitter, &logger, ws, console)
	quitCh := make(chan struct{})
	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-signalCh
		close(quitCh)
	}()

	err = service.Run(quitCh)
	if err != nil {
		log.Fatal(err)
	}
}
