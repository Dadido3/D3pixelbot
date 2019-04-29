package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type connectionPixelcanvasio struct {
	OnlinePlayers          uint32 // Must be read atomically
	GoroutineQueryQuit     chan struct{}
	GoroutineWebsocketQuit chan struct{}
}

type pixelcanvasioResponsePlayers struct {
	Online int `json:"online"`
}

type pixelcanvasioResponseWebsocketURL struct {
	URL string `json:"url"`
}

func newPixelcanvasio() (*connectionPixelcanvasio, error) {
	con := &connectionPixelcanvasio{
		GoroutineQueryQuit:     make(chan struct{}),
		GoroutineWebsocketQuit: make(chan struct{}),
	}

	// Main goroutine that handles queries and timed things
	go func(con *connectionPixelcanvasio) {
		queryTicker := time.NewTicker(10 * time.Second)
		defer queryTicker.Stop()

		getOnlinePlayers := func() {
			response := &pixelcanvasioResponsePlayers{}
			if err := getJSON("https://pixelcanvas.io/api/online", response); err == nil {
				atomic.StoreUint32(&con.OnlinePlayers, uint32(response.Online))
				log.Printf("Player amount: %v", response.Online)
			}
		}
		getOnlinePlayers()

		for {
			select {
			case <-queryTicker.C:
				getOnlinePlayers()
			case <-con.GoroutineQueryQuit:
				return
			}
		}
	}(con)

	// Main goroutine that handles the websocket connection (It will always try to reconnect)
	go func(con *connectionPixelcanvasio) {
		waitTime := 0 * time.Second
		for {
			select {
			case <-con.GoroutineWebsocketQuit:
				return
			case <-time.After(waitTime):
			}

			// Any following connection attempt should be delayed a few seconds
			waitTime = 5 * time.Second

			// Get websocket URL
			u, err := con.getWebsocketURL()
			if err != nil {
				log.Printf("Failed to connect to websocket server: %v", err)
				continue
			}

			u.RawQuery = "fingerprint=11111111111111111111111111111111"

			// Connect to websocket server
			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Printf("Failed to connect to websocket server %v: %v", u.String(), err)
				continue
			}

			// Wait for and handle external close events, or connection errors.
			quitChannel := make(chan struct{})
			go func(c *websocket.Conn, quitChannel chan struct{}) {
				select {
				case <-con.GoroutineWebsocketQuit:
					c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					select {
					case <-quitChannel:
					case <-time.After(time.Second):
					}
				case <-quitChannel:
				}
				c.Close()
			}(c, quitChannel)

			// Handle events
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Printf("Websocket connection error: %v", err)
					break
				}
				if len(message) >= 1 {
					opcode := uint8(message[0])
					switch opcode {
					case 0xC1:
						if len(message) == 7 {
							cx := int16(binary.BigEndian.Uint16(message[1:]))
							cy := int16(binary.BigEndian.Uint16(message[3:]))
							mixed := binary.BigEndian.Uint16(message[5:])
							colorIndex := int(mixed & 0x0F)
							ox := int((mixed >> 4) & 0x3F)
							oy := int((mixed >> 10) & 0x3F)
							log.Printf("Pixelchange: color %v @ chunk %v, %v with offset %v, %v", colorIndex, cx, cy, ox, oy)
						}
					default:
						log.Printf("Unknown websocket opcode: %v", opcode)
					}

				}
			}
			close(quitChannel)

		}
	}(con)

	return con, nil
}

func (con *connectionPixelcanvasio) getOnlinePlayers() int {
	return int(atomic.LoadUint32(&con.OnlinePlayers))
}

func (con *connectionPixelcanvasio) getWebsocketURL() (u *url.URL, err error) {
	response := &pixelcanvasioResponseWebsocketURL{}
	if err := getJSON("https://pixelcanvas.io/api/ws", response); err != nil {
		return nil, fmt.Errorf("Couldn't retrieve websocket URL: %v", err)
	}

	u, err = url.Parse(response.URL)
	if err != nil {
		return nil, fmt.Errorf("Retrieved invalid websocket URL: %v", err)
	}

	return u, nil
}

func (con *connectionPixelcanvasio) Close() {
	// Close channels to send the "done" signal
	close(con.GoroutineQueryQuit)
	close(con.GoroutineWebsocketQuit)

	return
}
