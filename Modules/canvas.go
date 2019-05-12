package main

// TODO: Logic to unload not needed chunks after time

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"sync"
)

type canvasEventSetPixel struct {
	Pos        image.Point
	ColorIndex uint8
}

type canvasEventInvalidateRect struct {
	Rect image.Rectangle
}

type canvasEventInvalidateAll struct{}

type canvasEventSetImage struct {
	Image *image.Paletted
}

// TODO: Add more events: revalidate (when just a few pixels have changed after redownloading/validating a chunk)

type canvasListener interface {
	handleSetPixel(pos image.Point, colorInde uint8) error
	handleInvalidateRect(rect image.Rectangle) error
	handleInvalidateAll() error
	handleSetImage(img *image.Paletted) error

	// TODO: Add listening rectangles, outgoing channel and more in here
	// TODO: Add a way so listeners can define a list of chunks or rects the canvas (the game client finally) has to keep valid
}

type canvas struct {
	sync.RWMutex

	Chunks map[chunkCoordinate]*chunk

	ChunkSize pixelSize
	Palette   color.Palette

	EventChan     chan interface{}        // Forwards incoming changes to the broadcaster goroutine
	GoroutineQuit chan struct{}           // Closing this channel stops the goroutines
	Listeners     map[canvasListener]bool // Events get forwarded to these listeners
}

func newCanvas(chunkSize pixelSize, palette color.Palette) *canvas {
	can := &canvas{
		Chunks:        make(map[chunkCoordinate]*chunk, 0),
		ChunkSize:     chunkSize,
		Palette:       palette,
		EventChan:     make(chan interface{}), // TODO: Determine optimal chan size
		GoroutineQuit: make(chan struct{}),
		Listeners:     make(map[canvasListener]bool),
	}

	// Goroutine that handles event broadcasting to listeners, and writes events to disk.
	go func() {
		for {
			select {
			case event := <-can.EventChan:
				switch event := event.(type) {
				case canvasEventSetPixel:
					for listener := range can.Listeners {
						listener.handleSetPixel(event.Pos, event.ColorIndex)
					}
				case canvasEventSetImage:
					for listener := range can.Listeners {
						listener.handleSetImage(event.Image)
					}
				case canvasEventInvalidateRect:
					for listener := range can.Listeners {
						listener.handleInvalidateRect(event.Rect)
					}
				case canvasEventInvalidateAll:
					for listener := range can.Listeners {
						listener.handleInvalidateAll()
					}
				default:
					log.Fatalf("Unknown event occured: %T", event)
				}
			case <-can.GoroutineQuit:
				return
			}
		}
	}()

	return can
}

func (can *canvas) subscribeListener(l canvasListener) {
	can.Lock()
	defer can.Unlock()

	can.Listeners[l] = true
}

func (can *canvas) unsubscribeListener(l canvasListener) {
	can.Lock()
	defer can.Unlock()

	delete(can.Listeners, l)
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
			can.Palette,
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
			chunks = append(chunks, chunk)
		}
	}

	return chunks, nil
}

func (can *canvas) getPixel(pos image.Point) (color.Color, error) {
	chunkCoord := can.ChunkSize.getChunkCoord(pos)

	chunk, err := can.getChunk(chunkCoord, false)
	if err != nil {
		return nil, fmt.Errorf("Cannot get pixel at %v: %v", pos, err)
	}

	return chunk.getPixel(pos)
}

func (can *canvas) getPixelIndex(pos image.Point) (uint8, error) {
	chunkCoord := can.ChunkSize.getChunkCoord(pos)

	chunk, err := can.getChunk(chunkCoord, false)
	if err != nil {
		return 0, fmt.Errorf("Cannot get pixel at %v: %v", pos, err)
	}

	return chunk.getPixelIndex(pos)
}

func (can *canvas) setPixel(pos image.Point, col color.Color) error {
	return can.setPixelIndex(pos, uint8(can.Palette.Index(col)))
}

func (can *canvas) setPixelIndex(pos image.Point, colorIndex uint8) error {
	chunkCoord := can.ChunkSize.getChunkCoord(pos)

	// Forward event to broadcaster goroutine
	can.EventChan <- canvasEventSetPixel{
		Pos:        pos,
		ColorIndex: colorIndex,
	}

	chunk, err := can.getChunk(chunkCoord, false)
	if err != nil {
		return fmt.Errorf("Cannot set pixel at %v: %v", pos, err)
	}

	return chunk.setPixelIndex(pos, colorIndex)
}

// Will update the canvas with the given image.
// Only chunks that are fully inside the image will be updated.
// Missing chunks will be created.
func (can *canvas) setImage(img image.Image) error {
	chunkRect := can.ChunkSize.getInnerChunkRect(img.Bounds())
	chunks, err := can.getChunks(chunkRect, true, false)
	if err != nil {
		return fmt.Errorf("Could not draw image at %v: %v", img.Bounds(), err)
	}

	for _, chunk := range chunks {
		resultImg, err := chunk.setImage(img)
		if err != nil {
			return fmt.Errorf("Could not draw image at %v: %v", img.Bounds(), err)
		}
		// Forward event to broadcaster goroutine
		can.EventChan <- canvasEventSetImage{
			Image: resultImg,
		}
	}

	return nil
}

// Get image of the given rectangle.
// The resulting image can be in an inconsistent state when some chunks change while it's generated.
// To get consistent updates, you should rather subscribe to the canvas change broadcast.
// Invalid or not existent chunks will be drawn with palette color 0.
func (can *canvas) getImageCopy(rect image.Rectangle) (*image.Paletted, error) {
	chunkRect := can.ChunkSize.getOuterChunkRect(rect)
	chunks, err := can.getChunks(chunkRect, false, true)
	if err != nil {
		return nil, fmt.Errorf("Could not grab image at %v: %v", rect, err)
	}

	img := image.NewPaletted(rect, can.Palette)

	for _, chunk := range chunks {
		draw.Draw(img, rect, chunk.getImageCopy(), rect.Min, draw.Over)
	}

	return img, nil
}

// Invalidates all chunks the rectangle intersects with.
// This will create new chunks if needed.
//
// This should be used to signal the redownloading of a specific area or chunk
func (can *canvas) invalidateRect(rect image.Rectangle) error {
	// Forward event to broadcaster goroutine
	can.EventChan <- canvasEventInvalidateRect{
		Rect: rect,
	}

	chunkRect := can.ChunkSize.getOuterChunkRect(rect)
	chunks, err := can.getChunks(chunkRect, true, true)
	if err != nil {
		return fmt.Errorf("Could not invalidate rectangle %v: %v", rect, err)
	}

	for _, chunk := range chunks {
		chunk.invalidateImage()
	}

	return nil
}

// Invalidates all chunks.
// This will only affect existing chunks.
//
// That should be used to signal connection loss or the redownloading of everything
func (can *canvas) invalidateAll() error {
	can.RLock()
	defer can.RUnlock()

	// Forward event to broadcaster goroutine
	can.EventChan <- canvasEventInvalidateAll{}

	for _, chunk := range can.Chunks {
		chunk.invalidateImage()
	}

	return nil
}

func (can *canvas) Close() {
	// Stop gorountines gracefully
	close(can.GoroutineQuit)

	// TODO: Wait until gorountines are stopped

	return
}
