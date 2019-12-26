package subscription

import (
	"github.com/gorilla/websocket"
	"net/url"
)

type Websocket struct {
	Topic  Topic
	URL    url.URL
	Done   chan struct{}
	Err    chan error
	Msg    chan []byte
	Client *websocket.Conn
}

func (w *Websocket) Subscribe() error {
	c, _, err := websocket.DefaultDialer.Dial(w.URL.String(), nil)
	if err != nil {
		return err
	}
	w.Client = c

	sub, err := w.Topic.Stringify()
	if err != nil {
		return err
	}

	err2 := c.WriteMessage(websocket.TextMessage, sub)
	if err2 != nil {
		return err2
	}

	return nil
}

func (w *Websocket) Start() {
	go func() {
		for {
			_, message, err := w.Client.ReadMessage()
			if err != nil {
				w.Err <- err
			}
			w.Msg <- message
		}
	}()

}
