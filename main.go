package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jcuga/golongpoll"
	"github.com/jcuga/golongpoll/client"
	"github.com/labstack/echo/v4"
)

func main() {
	// Create longpoll manger with default opts
	manager, err := golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		panic(err)
	}

	
	router := echo.New()
	router.GET("/get", wrapWithContext(manager.SubscriptionHandler))
	router.POST("/post", wrapWithContext(manager.PublishHandler))

	go getMessages(manager)

	router.Logger.Fatal(router.Start(":8081"))
}

func wrapWithContext(lpHandler func(http.ResponseWriter, *http.Request)) func(echo.Context) error {
	return func(c echo.Context) error {
		lpHandler(c.Response().Writer, c.Request())
		return nil
	}
}

func getMessages(lpManager *golongpoll.LongpollManager) {
	u, err := url.Parse("http://127.0.0.1:8081/get")
	if err != nil {
		panic(err)
	}

	c, err := client.NewClient(client.ClientOptions{
		SubscribeUrl:   *u,
		Category:       "messages",
		LoggingEnabled: true,
	})
	if err != nil {
		panic(err)
	}

	for event := range c.Start(time.Now()) {
		msg := event.Data.(string)
		normMsg := strings.ToLower(msg)
		
		lpManager.Publish("messages", normMsg)
	}
	
	fmt.Println("stopping")
}