package main

import (
	"image"
	"image/color"
	"image/draw"
	"sync"
)

type chunkCoordinate pixelCoordinate
type chunkSize chunkCoordinate

type pixelQueueElement struct {
	Pos        pixelCoordinate
	ColorIndex uint8
}

type chunk struct {
	sync.RWMutex

	Rect    image.Rectangle
	Palette color.Palette
	Image   *image.Paletted // TODO: Compress or unload image when not needed

	pixelQueue []pixelQueueElement
	Valid      bool
}

func newChunk(rect image.Rectangle, p color.Palette) (*chunk, error) {
	chunk := &chunk{
		Rect:       rect,
		Palette:    p,
		Image:      image.NewPaletted(rect, p),
		pixelQueue: []pixelQueueElement{},
	}

	return chunk, nil
}

func (chu *chunk) getPixel(pos pixelCoordinate) color.Color {
	chu.RLock()
	defer chu.RUnlock()

	return chu.Image.At(pos.X, pos.Y) // TODO: Make this call secure, it causes a runtime error when it tries to retrieve a index outside the palette.
}

func (chu *chunk) getPixelIndex(pos pixelCoordinate) uint8 {
	chu.RLock()
	defer chu.RUnlock()

	return chu.Image.ColorIndexAt(pos.X, pos.Y)
}

func (chu *chunk) setPixel(pos pixelCoordinate, col color.Color) {
	chu.setPixelIndex(pos, uint8(chu.Image.Palette.Index(col)))
}

func (chu *chunk) setPixelIndex(pos pixelCoordinate, colorIndex uint8) {
	chu.Lock()
	defer chu.Unlock()

	if chu.Valid {
		chu.Image.SetColorIndex(pos.X, pos.Y, colorIndex)
		return
	}

	// If chunk is not valid, append to queue to draw them later
	chu.pixelQueue = append(chu.pixelQueue, pixelQueueElement{
		Pos:        pos,
		ColorIndex: colorIndex,
	})
}

// Overwrites the image data, and validates the chunk.
// The image rectangle needs to be aligned with the chunk rectangle.
func (chu *chunk) setImage(img image.Image) {
	chu.Lock()
	defer chu.Unlock()

	draw.Draw(chu.Image, chu.Image.Rect, img, chu.Image.Rect.Min, draw.Over)

	return
}

func (chu *chunk) getImageCopy() image.Image {
	img := chu.getImageCopyPaletted()

	return &img
}

// Overwrites the image data, and validates the chunk.
func (chu *chunk) setImagePaletted(img image.Paletted) {
	chu.Lock()
	defer chu.Unlock()

	chu.Image = &img
	chu.Valid = true
	return
}

func (chu *chunk) getImageCopyPaletted() image.Paletted {
	chu.RLock()
	defer chu.RUnlock()

	img := *chu.Image
	copy(img.Pix, chu.Image.Pix)
	copy(img.Palette, chu.Image.Palette)

	return img
}

// Invalidates the image, which shows that this chunk contains older or completely wrong data.
// While a chunk is invalid, all setPixel() calls will be queued.
// The only way to make a chunk valid, is by using setImage().
func (chu *chunk) invalidateImage(image.RGBA) {
	chu.Lock()
	defer chu.Unlock()

	chu.Valid = false
	chu.pixelQueue = []pixelQueueElement{}

	return
}
