package main

import (
	"fmt"
	notificationcenter "github.com/fallenstedt/sse/notification_center"
	twitterstream "github.com/fallenstedt/twitter-stream"
	"log"
	"net/http"
)

func main() {
	tok, err := twitterstream.NewTokenGenerator().SetApiKeyAndSecret(
		"YOUR_TWITTER_API_KEY",
		"YOUR_TWITTER_API_SECRET",
		).RequestBearerToken()

	if err != nil {
		panic(err)
	}

	stream := twitterstream.NewTwitterStream(tok.AccessToken)
	err = stream.Stream.StartStream()

	if err != nil {
		panic(err)
	}

	nc := notificationcenter.NewNotificationCenter()

	go func() {

		for message := range *stream.Stream.GetMessages() {
			str := fmt.Sprintf("%v", message)
			m := []byte(str)
			if err := nc.Notify(m); err != nil {
				log.Fatal(err)
			}
		}
	}()

	http.HandleFunc("/sse", handleSSE(nc))
	log.Println("Starting server on port 8001")
	log.Fatal(http.ListenAndServe(":8001", nil))

}

func handleSSE(s notificationcenter.Subscriber) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("New connection made")
		// Subscribe
		c := make(chan []byte)
		unsubscribeFn, err := s.Subscribe(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Signal SSE Support
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")


		for {
			select {
			case <-r.Context().Done():
				if err := unsubscribeFn(); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				break

			default:
				b := <-c
				fmt.Fprintf(w, "data: %s\n\n", b)
				w.(http.Flusher).Flush()
			}
		}
	}
}
