/*  D3pixelbot - Custom client, recorder and bot for pixel drawing games
    Copyright (C) 2019  David Vogel

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.  */

// TODO: Send pixels to game API
// TODO: Handle captchas, and forward them somewhere

package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var pixelcanvasioChunkSize = pixelSize{64, 64} // Not the chunk size that the canvas is initialized with
var pixelcanvasioChunkCollectionRadius = 7
var pixelcanvasioChunkCollectionSize = chunkSize{pixelcanvasioChunkCollectionRadius*2 + 1, pixelcanvasioChunkCollectionRadius*2 + 1} // Arraysize of chunks that's returned on the bigchunk request
var pixelcanvasioChunkOffset = image.Point{pixelcanvasioChunkCollectionRadius * pixelcanvasioChunkSize.X, pixelcanvasioChunkCollectionRadius * pixelcanvasioChunkSize.Y}
var pixelcanvasioChunkCollectionPixelSize = pixelSize{pixelcanvasioChunkCollectionSize.X * pixelcanvasioChunkSize.X, pixelcanvasioChunkCollectionSize.Y * pixelcanvasioChunkSize.Y}
var pixelcanvasioCanvasRect = image.Rectangle{image.Point{-999999, -999999}, image.Point{1000000, 1000000}}

var pixelcanvasioPalette = []color.Color{
	color.RGBA{255, 255, 255, 255},
	color.RGBA{228, 228, 228, 255},
	color.RGBA{136, 136, 136, 255},
	color.RGBA{34, 34, 34, 255},
	color.RGBA{255, 167, 209, 255},
	color.RGBA{229, 0, 0, 255},
	color.RGBA{229, 149, 0, 255},
	color.RGBA{160, 106, 66, 255},
	color.RGBA{229, 217, 0, 255},
	color.RGBA{148, 224, 68, 255},
	color.RGBA{2, 190, 1, 255},
	color.RGBA{0, 211, 221, 255},
	color.RGBA{0, 131, 199, 255},
	color.RGBA{0, 0, 234, 255},
	color.RGBA{207, 110, 228, 255},
	color.RGBA{130, 0, 128, 255},
}

type connectionPixelcanvasio struct {
	RefCounter int

	Fingerprint      string
	OnlinePlayers    uint32 // Must be read atomically
	Center           image.Point
	AuthName, AuthID string
	NextPixel        time.Time

	Canvas *canvas

	GoroutineQuit     chan struct{} // Closing this channel stops the goroutines
	QuitWaitgroup     sync.WaitGroup
	ChunkDownloadChan <-chan *chunk // Receives download requests from the canvas
}

var pixelcanvasioConnection *connectionPixelcanvasio // TODO: Use/Create something similar to sync.Once, but with a counter
var pixelcanvasioConnectionMutex = &sync.Mutex{}

func init() {
	// Register connection types (all init functions are called from a single thread, thus threadsafe)
	connectionTypes["pixelcanvasio"] = connectionType{
		Name:        "PixelCanvas.io",
		FunctionNew: newPixelcanvasio,
	}
}

func newPixelcanvasio() (connection, *canvas) {
	pixelcanvasioConnectionMutex.Lock()
	defer pixelcanvasioConnectionMutex.Unlock()

	con := pixelcanvasioConnection

	if con == nil {
		con = &connectionPixelcanvasio{
			RefCounter:    1,
			Fingerprint:   "11111111111111111111111111111111",
			GoroutineQuit: make(chan struct{}),
		}
	} else {
		// Reuse connection if there is already an instance
		con.RefCounter++
		return con, con.Canvas
	}

	pixelcanvasioConnection = con

	con.Canvas, con.ChunkDownloadChan = newCanvas(pixelcanvasioChunkCollectionPixelSize, pixelcanvasioChunkOffset, pixelcanvasioCanvasRect)

	// Main goroutine that handles queries and timed things
	con.QuitWaitgroup.Add(1)
	go func() {
		defer con.QuitWaitgroup.Done()

		queryTicker := time.NewTicker(10 * time.Second)
		defer queryTicker.Stop()

		getOnlinePlayers := func() {
			response := &struct {
				Online int `json:"online"`
			}{}
			if err := getJSON("https://pixelcanvas.io/api/online", response); err == nil {
				atomic.StoreUint32(&con.OnlinePlayers, uint32(response.Online))
				log.Debugf("Player amount: %v", response.Online)
			}
		}
		getOnlinePlayers()

		for {
			select {
			case <-queryTicker.C:
				getOnlinePlayers()
			case <-con.GoroutineQuit:
				return
			}
		}
	}()

	myClient := &http.Client{Timeout: 1 * time.Minute}
	downloadWaitgroup := sync.WaitGroup{}   // To wait until all downloads are finished
	downloadLimit := make(chan struct{}, 3) // Limit maximum amount of simultaneous downloads to 3
	handleDownload := func(chu *chunk) error {
		// Round to nearest bigchunk // TODO: Simplify, especially as there is an origin parameter now
		ccOffset := image.Point(pixelcanvasioChunkSize).Mul(pixelcanvasioChunkCollectionRadius)
		cc := pixelcanvasioChunkCollectionSize.getPixelSize(pixelcanvasioChunkSize).getChunkCoord(chu.Rect.Min.Add(ccOffset), image.Point{})
		cc.X, cc.Y = cc.X*pixelcanvasioChunkCollectionSize.X, cc.Y*pixelcanvasioChunkCollectionSize.Y
		ca := chunkRectangle{image.Rectangle{
			Min: image.Point(cc).Add(image.Point{-pixelcanvasioChunkCollectionRadius, -pixelcanvasioChunkCollectionRadius}),
			Max: image.Point(cc).Add(image.Point{pixelcanvasioChunkCollectionRadius + 1, pixelcanvasioChunkCollectionRadius + 1}),
		}}.getPixelRectangle(pixelcanvasioChunkSize, image.Point{})

		// Signalling must not be in the goroutine, so that the download isn't started several times because of neighbors
		chunks, err := con.Canvas.signalDownload(ca)
		if err != nil {
			return fmt.Errorf("Can't signal downloading of chunks at %v: %v", cc, err)
		}
		if len(chunks) == 0 {
			return fmt.Errorf("Couldn't signal download for any chunk at %v", cc)
		}
		// TODO: Only setImage on chunks returned by signalDownload

		log.Tracef("Download at %v signalled", cc)

		downloadWaitgroup.Add(1)
		go func() {
			downloadLimit <- struct{}{} // Block inside the goroutine, so downloads will queue up without blocking anything else
			defer downloadWaitgroup.Done()
			defer func() { <-downloadLimit }()

			startTime := time.Now()
			log.Tracef("Download at %v started", cc)

			r, err := myClient.Get(fmt.Sprintf("https://pixelcanvas.io/api/bigchunk/%v.%v.bmp", cc.X, cc.Y))
			if err != nil {
				log.Errorf("Can't get bigchunk at %v: %v", cc, err)
				return
			}
			defer r.Body.Close()

			raw, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Errorf("Error in bigchunk result: %v", err)
				return
			}
			expectedLen := pixelcanvasioChunkSize.X * pixelcanvasioChunkSize.Y * ((pixelcanvasioChunkCollectionSize.X) * (pixelcanvasioChunkCollectionSize.Y)) / 2
			if len(raw) != expectedLen {
				log.Errorf("Returned image data has the wrong length (%v, expected %v)", len(raw), expectedLen)
				log.Errorf("API returned %v", string(raw[:1000]))
				return
			}

			downloadTime := time.Now().Sub(startTime).Seconds()
			startTime = time.Now()

			img := image.NewPaletted(ca, pixelcanvasioPalette)
			i := 0

			for iy := 0; iy < pixelcanvasioChunkCollectionSize.Y; iy++ {
				for ix := 0; ix < pixelcanvasioChunkCollectionSize.X; ix++ {
					c := chunkCoordinate{
						X: cc.X + ix - pixelcanvasioChunkCollectionRadius,
						Y: cc.Y + iy - pixelcanvasioChunkCollectionRadius,
					}
					for jy := 0; jy < pixelcanvasioChunkSize.Y; jy++ {
						for jx := 0; jx < pixelcanvasioChunkSize.X; jx += 2 {
							p := image.Point{
								X: c.X*pixelcanvasioChunkSize.X + jx,
								Y: c.Y*pixelcanvasioChunkSize.Y + jy,
							}

							img.SetColorIndex(p.X, p.Y, (raw[i]>>4)&0x0F) // TODO: Optimize image drawing for receiving
							img.SetColorIndex(p.X+1, p.Y, raw[i]&0x0F)
							i++
						}
					}
				}
			}

			drawTime := time.Now().Sub(startTime).Seconds()
			startTime = time.Now()

			err = con.Canvas.setImage(img, false, true)
			if err != nil {
				log.Warningf("Can't set image at %v: %v", img.Rect, err)
				return
			}

			setTime := time.Now().Sub(startTime).Seconds()
			log.Tracef("Times for %v: Download %.3fs, Drawing %.3fs, setImage() %.5fs ", cc, downloadTime, drawTime, setTime)

		}()

		return nil
	}

	// Main goroutine that handles the websocket connection (It will always try to reconnect)
	con.QuitWaitgroup.Add(1)
	go func() {
		defer con.QuitWaitgroup.Done()

		waitTime := 0 * time.Second
		for {
			select {
			case <-con.GoroutineQuit:
				return
			case <-time.After(waitTime):
			}

			// Any following connection attempt should be delayed a few seconds
			waitTime = 5 * time.Second

			// Get websocket URL
			u, err := con.getWebsocketURL()
			if err != nil {
				log.Errorf("Failed to connect to websocket server: %v", err)
				continue
			}

			u.RawQuery = "fingerprint=" + con.Fingerprint

			// Connect to websocket server
			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil) // TODO: Connecting pinging and timeouts
			if err != nil {
				log.Errorf("Failed to connect to websocket server %v: %v", u.String(), err)
				continue
			}

			// Handle chunk downloading in a goroutine
			chunkDownloaderQuit := make(chan struct{})
			go func() {
				for {
					select {
					case chu := <-con.ChunkDownloadChan:
						// Check if the chunk still needs to be downloaded
						if chu.getQueryState(false) == chunkDownload {
							handleDownload(chu)
						}
					case <-chunkDownloaderQuit:
						return
					}
				}
			}()

			// Wait for and handle external close events, or connection errors
			quitChannel := make(chan struct{})
			go func(c *websocket.Conn, quitChannel chan struct{}) {
				select {
				case <-con.GoroutineQuit:
					c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					select {
					case <-quitChannel:
					case <-time.After(time.Second):
					}
				case <-quitChannel:
				}
				c.Close()
			}(c, quitChannel)

			log.Debugf("Websocket connection opened")

			// Handle events
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Warnf("Websocket connection error: %v", err)
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
							colorIndex := uint8(mixed & 0x0F)
							color := pixelcanvasioPalette[colorIndex] // colorIndex technically can't be >= 16, so it should be save
							ox := int((mixed >> 4) & 0x3F)
							oy := int((mixed >> 10) & 0x3F)
							log.Tracef("Pixelchange: color %v @ chunk %v, %v with offset %v, %v", colorIndex, cx, cy, ox, oy)
							pos := image.Point{
								X: int(cx)*pixelcanvasioChunkSize.X + ox,
								Y: int(cy)*pixelcanvasioChunkSize.Y + oy,
							}
							if err := con.Canvas.setPixel(pos, color); err != nil {
								log.Debugf("Couldn't draw pixel at %v with color %v: %v", pos, colorIndex, err)
							}
						}
					default:
						log.Errorf("Unknown websocket opcode: %v", opcode)
					}

				}
			}
			log.Debugf("Websocket connection closed")
			close(chunkDownloaderQuit)
			close(quitChannel)
			downloadWaitgroup.Wait() // Wait until all chunk downloads are finished
			log.Tracef("All downloads finished")

			con.Canvas.invalidateAll()

		}
	}()

	// TODO: Authenticate before setting/sending a pixel
	//fmt.Print(con.authenticateMe())

	return con, con.Canvas
}

func (con *connectionPixelcanvasio) getShortName() string {
	return "pixelcanvasio"
}

func (con *connectionPixelcanvasio) getName() string {
	return "PixelCanvas.io"
}

func (con *connectionPixelcanvasio) getOnlinePlayers() int {
	return int(atomic.LoadUint32(&con.OnlinePlayers))
}

func (con *connectionPixelcanvasio) getWebsocketURL() (u *url.URL, err error) {
	response := &struct {
		URL string `json:"url"`
	}{}
	if err := getJSON("https://pixelcanvas.io/api/ws", response); err != nil {
		return nil, fmt.Errorf("Couldn't retrieve websocket URL: %v", err)
	}

	u, err = url.Parse(response.URL)
	if err != nil {
		return nil, fmt.Errorf("Retrieved invalid websocket URL: %v", err)
	}

	return u, nil
}

func (con *connectionPixelcanvasio) authenticateMe() error {
	// TODO: Make threadsafe
	request := struct {
		Fingerprint string `json:"fingerprint"`
	}{
		Fingerprint: con.Fingerprint,
	}

	statusCode, _, body, err := postJSON("https://pixelcanvas.io/api/me", "https://pixelcanvas.io/", request)
	if err != nil {
		return err
	}

	response := &struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Center      []int   `json:"center"`
		WaitSeconds float32 `json:"waitSeconds"`
	}{}
	if err := json.Unmarshal(body, response); err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("Authentication failed with wrong status code: %v (body: %v)", statusCode, string(body))
	}

	if len(response.Center) < 2 {
		return fmt.Errorf("Invalid center given in authentication response")
	}

	con.AuthID = response.ID
	con.AuthName = response.Name
	con.Center.X, con.Center.Y = response.Center[0], response.Center[1]
	con.NextPixel = time.Now().Add(time.Duration(response.WaitSeconds*1000) * time.Millisecond)

	return nil
}

// Closes connection and canvas
func (con *connectionPixelcanvasio) Close() {
	pixelcanvasioConnectionMutex.Lock()
	defer pixelcanvasioConnectionMutex.Unlock()

	con.RefCounter--
	if con.RefCounter > 0 {
		return
	}

	pixelcanvasioConnection = nil

	// Stop goroutines gracefully
	close(con.GoroutineQuit)

	log.Trace("Waiting for downloads to finish")
	con.QuitWaitgroup.Wait()

	con.Canvas.Close()

	return
}
