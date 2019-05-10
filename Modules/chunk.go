package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"sync"
)

type pixelQueueElement struct {
	Pos        image.Point
	ColorIndex uint8
}

type chunk struct {
	sync.RWMutex

	Rect    image.Rectangle
	Palette color.Palette
	Image   *image.Paletted // TODO: Compress or unload image when not needed

	PixelQueue []pixelQueueElement
	Valid      bool
}

func newChunk(rect image.Rectangle, p color.Palette) *chunk {
	chunk := &chunk{
		Rect:       rect.Canon(),
		Palette:    p,
		Image:      image.NewPaletted(rect, p),
		PixelQueue: []pixelQueueElement{},
	}

	return chunk
}

func (chu *chunk) getPixel(pos image.Point) (color.Color, error) {
	chu.RLock()
	defer chu.RUnlock()

	if !pos.In(chu.Rect) {
		return nil, fmt.Errorf("Position is outside of the chunk")
	}

	return chu.Image.At(pos.X, pos.Y), nil // TODO: Make this call secure, it causes a runtime error when it tries to retrieve an index outside the palette.
}

func (chu *chunk) getPixelIndex(pos image.Point) (uint8, error) {
	chu.RLock()
	defer chu.RUnlock()

	if !pos.In(chu.Rect) {
		return 0, fmt.Errorf("Position is outside of the chunk")
	}

	return chu.Image.ColorIndexAt(pos.X, pos.Y), nil
}

func (chu *chunk) setPixel(pos image.Point, col color.Color) error {
	return chu.setPixelIndex(pos, uint8(chu.Image.Palette.Index(col)))
}

func (chu *chunk) setPixelIndex(pos image.Point, colorIndex uint8) error {
	chu.Lock()
	defer chu.Unlock()

	if !pos.In(chu.Rect) {
		return fmt.Errorf("Position is outside of the chunk")
	}

	if chu.Valid {
		chu.Image.SetColorIndex(pos.X, pos.Y, colorIndex)
		return nil
	}

	// If chunk is not valid, append to queue to draw them later
	chu.PixelQueue = append(chu.PixelQueue, pixelQueueElement{
		Pos:        pos,
		ColorIndex: colorIndex,
	})

	return nil
}

// Overwrites the image data, and validates the chunk.
// The chunk boundaries needs to be inside the image boundaries, otherwise the operation will fail.
//
// All queued pixels will be replayed when this function is called.
// This helps to prevent inconsitencies while downloading chunks.
func (chu *chunk) setImage(img image.Image) error {
	chu.Lock()
	defer chu.Unlock()

	if !chu.Rect.In(img.Bounds()) {
		return fmt.Errorf("The chunk boundaries aren't a subset of the image boundaries")
	}

	draw.Draw(chu.Image, chu.Image.Rect, img, chu.Image.Rect.Min, draw.Over) // TODO: Check if the sp parameter is correct

	// Replay all the queued pixels
	for _, pqe := range chu.PixelQueue {
		chu.Image.SetColorIndex(pqe.Pos.X, pqe.Pos.Y, pqe.ColorIndex)
	}

	chu.PixelQueue = []pixelQueueElement{}
	chu.Valid = true

	return nil
}

func (chu *chunk) getImageCopy() *image.Paletted {
	chu.RLock()
	defer chu.RUnlock()

	img := *chu.Image
	copy(img.Pix, chu.Image.Pix)
	copy(img.Palette, chu.Image.Palette)

	return &img
}

// Invalidates the image, which shows that this chunk contains old or completely wrong data.
// While a chunk is invalid, all setPixel() calls will be queued.
// The only way to make a chunk valid, is by using setImage().
func (chu *chunk) invalidateImage() {
	chu.Lock()
	defer chu.Unlock()

	chu.PixelQueue = []pixelQueueElement{} // Empty queue on new invalidation.
	chu.Valid = false

	return
}
