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

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sync"
	"time"
)

type canvasEventInvalidateAll struct{}

type canvasEventInvalidateRect struct {
	Rect image.Rectangle
}

type canvasEventSetImage struct {
	Image image.Image
}

type canvasEventSetPixel struct {
	Pos   image.Point
	Color color.Color
}

type canvasEventSignalDownload struct {
	Rect image.Rectangle
}

// TODO: Add more events: revalidate (when just a few pixels have changed after redownloading/validating a chunk)

type canvasEventListenerSubscribe struct {
	Listener canvasListener
}

type canvasEventListenerUnsubscribe struct {
	Listener canvasListener
}

type canvasEventListenerRects struct {
	Listener   canvasListener
	Rects      []image.Rectangle
	ForwardAll bool
}

type canvasListener interface {
	handleChunksChange(create, remove []image.Rectangle) error

	handleInvalidateAll() error
	handleInvalidateRect(rect image.Rectangle) error
	handleSetImage(img image.Image) error
	handleSetPixel(pos image.Point, color color.Color) error
	handleSignalDownload(rect image.Rectangle) error
}

type canvasListenerState struct {
	Rects      []image.Rectangle        // Rectangles that the listener needs to be kept up to do date with. The canvas will keep those rectangles in sync with the game
	Chunks     map[image.Rectangle]bool // Chunks rectangles the the listener knows
	ForwardAll bool                     // True: the listeners wants to get all events, even the ones outside his rectangles
}

type canvas struct {
	sync.RWMutex
	Closed      bool
	ClosedMutex sync.RWMutex

	Rect   image.Rectangle // Valid area of the canvas // TODO: Enforce canvas limit
	Chunks map[chunkCoordinate]*chunk

	ChunkSize pixelSize
	Palette   color.Palette

	EventChan        chan interface{} // Forwards incoming events to the goroutine
	ChunkRequestChan chan *chunk      // Chunk download requests that go to the game connection // TODO: Convert it to a method, not a channel
}

func newCanvas(chunkSize pixelSize, canvasRect image.Rectangle, palette color.Palette) (*canvas, <-chan *chunk) {
	can := &canvas{
		Chunks:           make(map[chunkCoordinate]*chunk, 0),
		ChunkSize:        chunkSize,
		Palette:          palette,
		EventChan:        make(chan interface{}), // TODO: Determine optimal chan size (Add waitGroup when channel buffering is enabled!)
		ChunkRequestChan: make(chan *chunk),
	}

	handleChunk := func(chunk *chunk, resetTime bool) {
		switch chunk.getQueryState(resetTime) {
		case chunkDelete:
			can.Lock()
			delete(can.Chunks, can.ChunkSize.getChunkCoord(chunk.Rect.Min))
			can.Unlock()
		case chunkDownload:
			can.ChunkRequestChan <- chunk
		}
	}

	rectQueryChan := make(chan image.Rectangle)

	// Goroutine that handles chunk downloading (Queries the game connection for chunks)
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case rect, ok := <-rectQueryChan:
				if !ok {
					// Close goroutine, as the channel is gone
					return
				}
				chunkRect := can.ChunkSize.getOuterChunkRect(rect)
				chunks, err := can.getChunks(chunkRect, true, true)
				if err == nil {
					for _, chunk := range chunks {
						handleChunk(chunk, true)
					}
				}
			case <-ticker.C: // Query all chunks for state changes every minute
				// Make copy of the chunks map
				can.RLock()
				chunks := make(map[chunkCoordinate]*chunk)
				for k, v := range can.Chunks {
					chunks[k] = v
				}
				can.RUnlock()
				for _, chunk := range chunks {
					handleChunk(chunk, false) // Handle chunks, but don't reset their timer
				}

			}
		}
	}()

	// Goroutine that handles event broadcasting to listeners
	// It can directly broadcast events from the EventChan, or it can create new events for specific listeners.
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		listeners := map[canvasListener]canvasListenerState{} // Events get forwarded to these listeners
		defer close(rectQueryChan)

		for {
			select {
			case event, ok := <-can.EventChan:
				if !ok {
					// Close goroutine, as the channel is gone
					log.Trace("Broadcaster closed!")
					return
				}
				switch event := event.(type) {
				case canvasEventSetPixel:
					for listener := range listeners { // TODO: Limit forwarding to rects
						listener.handleSetPixel(event.Pos, event.Color)
					}
				case canvasEventSetImage:
					for listener := range listeners {
						listener.handleSetImage(event.Image)
					}
				case canvasEventInvalidateRect:
					for listener := range listeners {
						listener.handleInvalidateRect(event.Rect)
					}
				case canvasEventInvalidateAll:
					for listener := range listeners {
						listener.handleInvalidateAll()
					}
				case canvasEventSignalDownload:
					for listener := range listeners {
						listener.handleSignalDownload(event.Rect)
					}
				case canvasEventListenerSubscribe:
					//log.Tracef("Listener %v subscribed", event.Listener)
					listeners[event.Listener] = canvasListenerState{}
				case canvasEventListenerUnsubscribe:
					//log.Tracef("Listener %v unsubscribed", event.Listener)
					delete(listeners, event.Listener)
				case canvasEventListenerRects:
					state, ok := listeners[event.Listener]
					if ok {
						//log.Tracef("Listener %v changed rects to %v", event.Listener, event.Rects)

						state.Rects = event.Rects
						state.ForwardAll = event.ForwardAll

						// Get all chunk rects that are intersecting the listener rectangles. Also query rects
						neededChunks := map[image.Rectangle]bool{}
						for _, rect := range event.Rects {
							go func(rect image.Rectangle) { rectQueryChan <- rect }(rect) // Async download request
							chunkRect := can.ChunkSize.getOuterChunkRect(rect)
							for iy := chunkRect.Min.Y; iy < chunkRect.Max.Y; iy++ {
								for ix := chunkRect.Min.X; ix < chunkRect.Max.X; ix++ {
									chunkRect := image.Rectangle{
										Min: image.Point{ix * can.ChunkSize.X, iy * can.ChunkSize.Y},
										Max: image.Point{(ix + 1) * can.ChunkSize.X, (iy + 1) * can.ChunkSize.Y},
									}
									neededChunks[chunkRect] = true
								}
							}
						}

						// Handle chunk rects, that are missing on the listeners side
						createChunks := []image.Rectangle{}
						for k := range neededChunks {
							if _, ok := state.Chunks[k]; !ok {
								createChunks = append(createChunks, k)
							}
						}

						// Handle chunk rects, that are not needed anymore on the listeners side
						removeChunks := []image.Rectangle{}
						for k := range state.Chunks {
							if _, ok := neededChunks[k]; !ok {
								removeChunks = append(removeChunks, k)
							}
						}

						state.Chunks = neededChunks
						listeners[event.Listener] = state

						event.Listener.handleChunksChange(createChunks, removeChunks)

						// Additionally send images for the new chunks if possible
						for _, rect := range createChunks {
							img, err := can.getImageCopy(rect, true, false)
							if err == nil {
								// Existing chunk with valid data, simulate download process to that specific listener
								event.Listener.handleSignalDownload(rect)
								event.Listener.handleSetImage(img) // TODO: Don't call that handler several times, especially sciter is slow in this case
							}
							// TODO: Send invalid chunks somehow. Maybe with a new handler
						}

					}
				default:
					log.Fatalf("Unknown event occurred: %T", event)
				}
			case <-ticker.C: // Query all rects every minute
				for _, state := range listeners {
					for _, rect := range state.Rects {
						go func(rect image.Rectangle) { rectQueryChan <- rect }(rect) // Async download request
					}
				}
			}
		}
	}()

	return can, can.ChunkRequestChan
}

func (can *canvas) subscribeListener(l canvasListener) error {
	can.ClosedMutex.RLock()
	defer can.ClosedMutex.RUnlock()
	if can.Closed {
		return fmt.Errorf("Canvas is closed")
	}

	// Forward event to broadcaster goroutine, even if there isn't a chunk.
	can.EventChan <- canvasEventListenerSubscribe{
		Listener: l,
	}

	return nil
}

func (can *canvas) unsubscribeListener(l canvasListener) error {
	can.ClosedMutex.RLock()
	defer can.ClosedMutex.RUnlock()
	if can.Closed {
		return fmt.Errorf("Canvas is closed")
	}

	// Forward event to broadcaster goroutine, even if there isn't a chunk.
	can.EventChan <- canvasEventListenerUnsubscribe{
		Listener: l,
	}

	return nil
}

// Register a number of rectangles that the listener needs to be kept up to date with.
// If forwardAll is true, any event is forwarded to the listener, even if it is outside the given rectangles.
//
// This function will not fail if the listener isn't subscribed.
//
// Don't call this function from the same context that handles events, or it will cause a deadlock.
func (can *canvas) registerRects(l canvasListener, rects []image.Rectangle, forwardAll bool) error {
	can.ClosedMutex.RLock()
	defer can.ClosedMutex.RUnlock()
	if can.Closed {
		return fmt.Errorf("Canvas is closed")
	}

	// Forward event to broadcaster goroutine, even if there isn't a chunk.
	can.EventChan <- canvasEventListenerRects{
		Listener:   l,
		Rects:      rects,
		ForwardAll: forwardAll,
	}

	return nil
}

func (can *canvas) getChunk(coord chunkCoordinate, createIfNonexistent bool) (*chunk, error) {
	if createIfNonexistent {
		can.Lock()
		defer can.Unlock()
	} else {
		can.RLock()
		defer can.RUnlock()
	}

	chunk, ok := can.Chunks[coord]
	if ok {
		return chunk, nil
	}

	if createIfNonexistent {
		min := image.Point{coord.X * can.ChunkSize.X, coord.Y * can.ChunkSize.Y}
		max := min.Add(image.Point{can.ChunkSize.X, can.ChunkSize.Y})
		chunk := newChunk(
			image.Rectangle{
				Min: min,
				Max: max,
			},
		)

		can.Chunks[coord] = chunk

		return chunk, nil
	}

	return nil, fmt.Errorf("Chunk at %v does not exist", coord)
}

func (can *canvas) getChunks(rect chunkRectangle, createIfNonexistent, ignoreNonexistent bool) ([]*chunk, error) {
	rectTemp := rect.Canon()
	chunks := []*chunk{}

	for iy := rectTemp.Min.Y; iy < rectTemp.Max.Y; iy++ {
		for ix := rectTemp.Min.X; ix < rectTemp.Max.X; ix++ {
			chunk, err := can.getChunk(chunkCoordinate{ix, iy}, createIfNonexistent)
			if err != nil && ignoreNonexistent == false {
				// This assumes that there can only be an error when createIfNonexistent == false
				// So it will never abort while it creates missing chunks
				return nil, fmt.Errorf("Can't get all chunks: %v", err)
			}
			if chunk != nil {
				chunks = append(chunks, chunk)
			}
		}
	}

	return chunks, nil
}

func (can *canvas) getPixel(pos image.Point) (color.Color, error) {
	chunkCoord := can.ChunkSize.getChunkCoord(pos)

	chunk, err := can.getChunk(chunkCoord, false)
	if err != nil {
		return nil, fmt.Errorf("Can't get chunk at %v: %v", pos, err)
	}

	return chunk.getPixel(pos)
}

func (can *canvas) getPixelIndex(pos image.Point) (uint8, error) {
	chunkCoord := can.ChunkSize.getChunkCoord(pos)

	chunk, err := can.getChunk(chunkCoord, false)
	if err != nil {
		return 0, fmt.Errorf("Can't get chunk at %v: %v", pos, err)
	}

	return chunk.getPixelIndex(pos)
}

func (can *canvas) setPixel(pos image.Point, col color.Color) error {
	can.ClosedMutex.RLock()
	defer can.ClosedMutex.RUnlock()
	if can.Closed {
		return fmt.Errorf("Canvas is closed")
	}

	// Forward event to broadcaster goroutine, even if there isn't a chunk. But send it after the chunk has been updated
	defer func() {
		can.EventChan <- canvasEventSetPixel{
			Pos:   pos,
			Color: col,
		}
	}()

	chunkCoord := can.ChunkSize.getChunkCoord(pos)

	chunk, err := can.getChunk(chunkCoord, false)
	if err != nil {
		return fmt.Errorf("Can't get chunk at %v: %v", chunkCoord, err)
	}

	return chunk.setPixel(pos, col)
}

// Will update the canvas with the given image.
// Only chunks that are fully inside the image will be updated.
// Chunks that have their download flag not set, will be ignored.
//
// This will validate the chunks, reset their download flag and replay any pixel events that happened while downloading.
// createIfNonexistent should be set to false normally.
func (can *canvas) setImage(img image.Image, createIfNonexistent, ignoreNonexistent bool) error {
	can.ClosedMutex.RLock()
	defer can.ClosedMutex.RUnlock()
	if can.Closed {
		return fmt.Errorf("Canvas is closed")
	}

	chunkRect := can.ChunkSize.getInnerChunkRect(img.Bounds())
	chunks, err := can.getChunks(chunkRect, createIfNonexistent, ignoreNonexistent)
	if err != nil {
		return fmt.Errorf("Can't get chunks from rectangle %v: %v", img.Bounds(), err)
	}

	for _, chunk := range chunks {
		resultImg, err := chunk.setImage(img)
		if err != nil {
			//return fmt.Errorf("Could not draw image at %v: %v", img.Bounds(), err)
			continue
		}
		// Forward event to broadcaster goroutine. It needs to be sent after chunk manipulation to keep everything in sync
		can.EventChan <- canvasEventSetImage{
			Image: resultImg,
		}
	}

	return nil
}

// Get RGBA image of the given rectangle.
// The resulting image can be in an inconsistent state when some chunks change while it's generated.
// But each chunk itself will be consistent.
// To get consistent updates, you should rather subscribe to the canvas change broadcast.
// If ignoreNonexistent is set to true, non existent chunks will be drawn transparent.
// If onlyIfValid is set to true, the function will fail if there are invalid chunks inside.
// If onlyIfValid is set to false, invalid chunks will be drawn transparent or with older data.
func (can *canvas) getImageCopy(rect image.Rectangle, onlyIfValid, ignoreNonexistent bool) (*image.RGBA, error) {
	chunkRect := can.ChunkSize.getOuterChunkRect(rect)
	chunks, err := can.getChunks(chunkRect, false, ignoreNonexistent)
	if err != nil {
		return nil, fmt.Errorf("Can't get chunks from rectangle %v: %v", rect, err)
	}

	img := image.NewRGBA(rect)

	for _, chunk := range chunks {
		imgCopy, err := chunk.getImageCopy(onlyIfValid)
		if err == nil {
			draw.Draw(img, rect, imgCopy, rect.Min, draw.Over)
		} else if onlyIfValid {
			return nil, fmt.Errorf("Can't get chunk image at %v: %v", chunk.Rect, err)
		}
	}

	return img, nil
}

// Invalidates all chunks the rectangle intersects with.
// This will only affect existing chunks.
//
// This should be used to signal connection loss or something that caused specific chunks to go out of sync.
func (can *canvas) invalidateRect(rect image.Rectangle) error {
	can.ClosedMutex.RLock()
	defer can.ClosedMutex.RUnlock()
	if can.Closed {
		return fmt.Errorf("Canvas is closed")
	}

	// Forward event to broadcaster goroutine. But send after chunks have been invalidated
	defer func() {
		can.EventChan <- canvasEventInvalidateRect{
			Rect: rect,
		}
	}()

	chunkRect := can.ChunkSize.getOuterChunkRect(rect)
	chunks, err := can.getChunks(chunkRect, false, true)
	if err != nil {
		return fmt.Errorf("Can't get chunks from rectangle %v: %v", rect, err)
	}

	for _, chunk := range chunks {
		chunk.invalidateImage()
	}

	return nil
}

// Invalidates all chunks.
// This will only affect existing chunks.
//
// This should be used to signal connection loss.
func (can *canvas) invalidateAll() error {
	can.ClosedMutex.RLock()
	defer can.ClosedMutex.RUnlock()
	if can.Closed {
		return fmt.Errorf("Canvas is closed")
	}

	can.RLock()
	defer can.RUnlock()

	for _, chunk := range can.Chunks {
		chunk.invalidateImage()
	}

	// Forward event to broadcaster goroutine
	can.EventChan <- canvasEventInvalidateAll{}

	return nil
}

// Returns true if the all intersecting chunks are valid and existent
func (can *canvas) isValid(rect image.Rectangle) bool {
	chunkRect := can.ChunkSize.getOuterChunkRect(rect)
	chunks, err := can.getChunks(chunkRect, false, false)
	if err != nil {
		return false
	}

	for _, chunk := range chunks {
		if !chunk.Valid {
			return false
		}
	}

	return true
}

// Signals that the specified rect is being downloaded.
// This will create new chunks if needed.
//
// A list of affected chunks is returned.
//
// This should be used to signal that the download for a specific area has started.
// A chunk that is in the downloading state will queue all pixel events, and will replay them after the download has finished.
// By replaying the pixels, the chunk will always be in sync with the game, even if downloading takes a while.
//
// For some game APIs it may not be necessary, as they send data serially.
// But signalDownload() must always be used, because otherwise the canvas would retrigger the download several times in a row on an invalid chunk.
func (can *canvas) signalDownload(rect image.Rectangle) ([]*chunk, error) {
	can.ClosedMutex.RLock()
	defer can.ClosedMutex.RUnlock()
	if can.Closed {
		return nil, fmt.Errorf("Canvas is closed")
	}

	// Forward event to broadcaster goroutine. But send after chunks have been flagged
	defer func() {
		can.EventChan <- canvasEventSignalDownload{
			Rect: rect,
		}
	}()

	chunkRect := can.ChunkSize.getOuterChunkRect(rect)
	chunks, err := can.getChunks(chunkRect, true, true)
	if err != nil {
		return nil, fmt.Errorf("Can't get chunks from rectangle %v: %v", rect, err)
	}

	downloading := []*chunk{}

	for _, chunk := range chunks {
		if chunk.signalDownload() {
			downloading = append(downloading, chunk)
		}
	}

	return downloading, nil
}

func (can *canvas) Close() {
	can.ClosedMutex.RLock()
	can.Closed = true // Prevent any new events from happening
	can.ClosedMutex.RUnlock()

	close(can.EventChan) // This will stop the goroutine after all events are processed

	return
}
