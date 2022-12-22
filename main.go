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
	router.POST("/post", wrapWithContextPost(manager.PublishHandler))

	go getMessages(manager)

	router.Logger.Fatal(router.Start(":8081"))
}

func wrapWithContext(lpHandler func(http.ResponseWriter, *http.Request)) func(echo.Context) error {
	return func(c echo.Context) error {
		lpHandler(c.Response().Writer, c.Request())
		return nil
	}
}

func wrapWithContextPost(lpHandler func(http.ResponseWriter, *http.Request)) func(echo.Context) error {
	return func(c echo.Context) error {
		fmt.Println("insert messages")
		lpHandler(c.Response().Writer, c.Request())
		return nil
	}
}

type requestBody struct {
	text string `json:"text"`
}

func getMessages(lpManager *golongpoll.LongpollManager) {
	u, err := url.Parse("http://127.0.0.1:8081/get")
	if err != nil {
		panic(err)
	}

	c, err := client.NewClient(client.ClientOptions{
		SubscribeUrl:   *u,
		Category:       "to-messages",
		LoggingEnabled: true,
	})
	if err != nil {
		panic(err)
	}

	for event := range c.Start(time.Now()) {
		req := event.Data.(map[string]interface{})

		text := req["text"].(string)
		normMsg := strings.ToLower(text)

		fmt.Println("select messages")
		fmt.Println(normMsg)

		lpManager.Publish("from-messages", normMsg)
	}

	fmt.Println("stopping")
}
