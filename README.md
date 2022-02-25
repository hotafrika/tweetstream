![example workflow](https://github.com/hotafrika/tweetstream/actions/workflows/autotests.yml/badge.svg)

This project is a tweets streaming engine. Original tweets source is Twitter's v2 Filtered Streaming API.
To use Twitter source of tweets you have to configure filtering rules.

Also there are a few tweets consumers implemented: websocket and console output. 
It is also possible to add any kind of consumer. It just has to satisfy Consumer interface.

At the time you can use more than one consumer.

Example usage
```go
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
```

