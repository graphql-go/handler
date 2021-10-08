package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"graphql-ws"},
}

type connectionACKMessage struct {
	OperationID string `json:"id,omitempty"`
	Type        string `json:"type"`
	Payload     struct {
		Query string `json:"query"`
	} `json:"payload,omitempty"`
}

type SubscriptionHandler struct {
	Schema *graphql.Schema
}

type Subscriber struct {
	Conn          *websocket.Conn
	RequestString string
	OperationID   string
}

func NewSubscriptionHandler(schema *graphql.Schema) *SubscriptionHandler {
	return &SubscriptionHandler{Schema: schema}
}

func (h *SubscriptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	connectionACK, err := json.Marshal(map[string]string{
		"type": "connection_ack",
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, connectionACK); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go h.handleSubscription(conn)
}

func (h *SubscriptionHandler) handleSubscription(conn *websocket.Conn) {
	var subscriber *Subscriber
	subscriptionCtx, subscriptionCancelFn := context.WithCancel(context.Background())

	handleClosedConnection := func() {
		h.unsubscribe(subscriptionCancelFn, subscriber)
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg connectionACKMessage
		if err := json.Unmarshal(p, &msg); err != nil {
			continue
		}

		if msg.Type == "stop" {
			handleClosedConnection()
			return
		}

		if msg.Type == "start" {
			subscriber = h.subscribe(subscriptionCtx, subscriptionCancelFn, conn, msg)
		}
	}
}

func (h *SubscriptionHandler) subscribe(ctx context.Context, cancelCtx context.CancelFunc, conn *websocket.Conn, msg connectionACKMessage) *Subscriber {
	subscriber := &Subscriber{
		Conn:          conn,
		RequestString: msg.Payload.Query,
		OperationID:   msg.OperationID,
	}

	sendMessage := func(r *graphql.Result) error {
		message, err := json.Marshal(map[string]interface{}{
			"type":    "data",
			"id":      subscriber.OperationID,
			"payload": r.Data,
		})
		if err != nil {
			return err
		}

		if err := subscriber.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return err
		}

		return nil
	}

	go func() {
		subscribeParams := graphql.Params{
			Context:       ctx,
			RequestString: msg.Payload.Query,
			Schema:        *h.Schema,
		}

		subscribeChannel := graphql.Subscribe(subscribeParams)

		for {
			select {
			case <-ctx.Done():
				return
			case r, isOpen := <-subscribeChannel:
				if !isOpen {
					h.unsubscribe(cancelCtx, subscriber)
					return
				}

				if err := sendMessage(r); err != nil {
					if err == websocket.ErrCloseSent {
						h.unsubscribe(cancelCtx, subscriber)
					}
				}
			}
		}
	}()

	return subscriber
}

func (h *SubscriptionHandler) unsubscribe(subscriptionCancelFn context.CancelFunc, subscriber *Subscriber) {
	subscriptionCancelFn()

	if subscriber != nil {
		subscriber.Conn.Close()
	}
}
