package client

import (
	"encoding/json"
	"log"
	"net/http"

	"kf/book"
	"kf/client/api"

	"github.com/gorilla/websocket"
)

type Kraken struct {
	book      *book.Book
	conn      *websocket.Conn
	api       api.API
	orderChan chan book.Order
}

func New(auth api.Auth, sig string, orderChan chan book.Order) *Kraken {
	return &Kraken{api: api.New(auth, sig), orderChan: orderChan}
}

func (kraken *Kraken) Close() {
	kraken.conn.Close()
	close(kraken.orderChan)
}

func (kraken *Kraken) Start(product string) {
	var err error
	kraken.conn, _, err = websocket.DefaultDialer.Dial(api.WsURL, http.Header{})
	if err != nil {
		log.Fatal(err)
	}

	var initResp map[string]interface{}
	kraken.conn.WriteMessage(websocket.TextMessage, []byte("{\"event\": \"subscribe\", \"feed\": \"book\", \"product_ids\": [\"" + product + "\"]}"))
	err = kraken.conn.ReadJSON(&initResp)
	if err != nil {
		log.Fatal(err)
	}

	err = kraken.conn.ReadJSON(&initResp)
	if err != nil {
		log.Fatal(err)
	} else if initResp["event"] != "subscribed" {
		log.Fatal(initResp)
	}

	var snapshot map[string]interface{}
	err = kraken.conn.ReadJSON(&snapshot)
	if err != nil {
		log.Fatal(err)
	}

	kraken.book = book.New(snapshot)
	go kraken.loop()
}

func (kraken *Kraken) loop() {
	for {
		_, delta, err := kraken.conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}

		var jsonDelta map[string]interface{}
		json.Unmarshal(delta, &jsonDelta)
		order := kraken.book.Update(jsonDelta)
		kraken.orderChan <- order
	}
}
