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

	"github.com/google/go-cmp/cmp"
)

const (
	chunkDeleteTime = 5 * time.Minute
)

type pixelQueueElement struct {
	Pos   image.Point
	Color color.Color
}

type chunk struct {
	sync.RWMutex

	Rect  image.Rectangle
	Image image.Image // TODO: Compress or unload image when not needed

	PixelQueue         []pixelQueueElement // Queued pixels, that are set while the image is downloading
	Valid, Downloading bool                // Valid: Data is in sync with the game. Downloading: Data is being downloaded. Both flags can't be true at the same time
	LastQueryTime      time.Time           // Point in time, when that chunk was queried last time. If this chunk hasn't been queried for some period, it will be unloaded.
}

// Create new empty chunk with rect
func newChunk(rect image.Rectangle) *chunk {
	cRect := rect.Canon()

	chunk := &chunk{
		Rect:          cRect,
		Image:         &cRect,
		PixelQueue:    []pixelQueueElement{},
		LastQueryTime: time.Now(),
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

	img, ok := chu.Image.(*image.Paletted)
	if !ok {
		return 0, fmt.Errorf("Chunk is not paletted")
	}

	return img.ColorIndexAt(pos.X, pos.Y), nil
}

func (chu *chunk) setPixel(pos image.Point, col color.Color) error {
	chu.Lock()
	defer chu.Unlock()

	if !pos.In(chu.Rect) {
		return fmt.Errorf("Position is outside of the chunk")
	}

	// TODO: Check if colorIndex is valid

	if chu.Valid {
		switch img := chu.Image.(type) {
		case *image.RGBA:
			img.Set(pos.X, pos.Y, col)
		case *image.Paletted:
			img.Set(pos.X, pos.Y, col)
		default:
			return fmt.Errorf("Incompatible chunk image type %T", img)
		}
	}

	// If chunk is downloading, append to queue to draw them later
	if chu.Downloading {
		chu.PixelQueue = append(chu.PixelQueue, pixelQueueElement{
			Pos:   pos,
			Color: col,
		})
	}

	return nil
}

func (chu *chunk) setPixelIndex(pos image.Point, colorIndex uint8) error {
	chu.Lock()
	defer chu.Unlock()

	if !pos.In(chu.Rect) {
		return fmt.Errorf("Position is outside of the chunk")
	}

	img, ok := chu.Image.(*image.Paletted)
	if !ok {
		return fmt.Errorf("Chunk is not paletted")
	}

	if int(colorIndex) >= len(img.Palette) {
		return fmt.Errorf("Color index outside of available palette")
	}
	img.SetColorIndex(pos.X, pos.Y, colorIndex)

	// If chunk is downloading, append to queue to draw them later
	if chu.Downloading {
		chu.PixelQueue = append(chu.PixelQueue, pixelQueueElement{
			Pos:   pos,
			Color: img.Palette[colorIndex],
		})
	}

	return nil
}

// Overwrites the image data, validates the chunk and resets the downloading flag.
// The chunk boundaries need to be inside the image boundaries, otherwise the operation will fail.
// Also, the download flag has to be set prior by using signalDownload().
//
// All queued pixels will be replayed when this function is called.
// This helps to prevent inconsistencies while downloading chunks.
// The result image is an up to date copy containing all queued changes.
func (chu *chunk) setImage(srcImg image.Image) (image.Image, error) {
	chu.Lock()
	defer chu.Unlock()

	if !chu.Rect.In(srcImg.Bounds()) {
		return nil, fmt.Errorf("The image doesn't fill the chunk completely")
	}
	if chu.Downloading == false {
		return nil, fmt.Errorf("The download flag isn't set")
	}

	// Create copy of a part of the source image. Fallback to RGBA image, if type is unknown
	switch srcImg := srcImg.(type) {
	case *image.Paletted:
		newImg := image.NewPaletted(chu.Rect, srcImg.Palette)
		draw.Draw(newImg, chu.Rect, srcImg, chu.Rect.Min, draw.Over)

		// If images are equal, copy nothing
		if cmp.Equal(chu.Image, newImg) && len(chu.PixelQueue) == 0 {
			chu.PixelQueue = []pixelQueueElement{}
			chu.Downloading = false
			chu.Valid = true

			return nil, nil // Return no image copy, this will cause the canvas to send a revalidate event
		}

		chu.Image = newImg
	default:
		newImg := image.NewRGBA(chu.Rect)
		draw.Draw(newImg, chu.Rect, srcImg, chu.Rect.Min, draw.Over)

		// If images are equal, copy nothing
		if cmp.Equal(chu.Image, newImg) && len(chu.PixelQueue) == 0 { // TODO: Compare seems slow, check it
			chu.PixelQueue = []pixelQueueElement{}
			chu.Downloading = false
			chu.Valid = true

			return nil, nil // Return no image copy, this will cause the canvas to send a revalidate event
		}

		chu.Image = newImg
	}

	// Replay all the queued pixels
	for _, pqe := range chu.PixelQueue {
		switch img := chu.Image.(type) {
		case *image.RGBA:
			img.Set(pqe.Pos.X, pqe.Pos.Y, pqe.Color)
		case *image.Paletted:
			img.Set(pqe.Pos.X, pqe.Pos.Y, pqe.Color)
		default:
			return nil, fmt.Errorf("Incompatible chunk image type %T", img)
		}
	}

	chu.PixelQueue = []pixelQueueElement{}
	chu.Downloading = false
	chu.Valid = true

	cpyImg, err := copyImage(chu.Image)
	if err != nil {
		return nil, fmt.Errorf("Couldn't copy image: %v", err)
	}

	return cpyImg, nil
}

func (chu *chunk) getImageCopy(onlyIfValid bool) (image.Image, bool, bool, error) {
	chu.RLock()
	defer chu.RUnlock()

	if onlyIfValid && !chu.Valid {
		return nil, false, false, fmt.Errorf("Chunk is not valid")
	}

	cpyImg, err := copyImage(chu.Image)
	if err != nil {
		return nil, false, false, fmt.Errorf("Couldn't copy image: %v", err)
	}

	return cpyImg, chu.Valid, chu.Downloading, nil
}

// Invalidates the image, which shows that this chunk contains old or completely wrong data.
//
// setImage() or revalidate() has to be used to signal that the chunk is valid again (in sync with the game).
func (chu *chunk) invalidateImage() {
	chu.Lock()
	defer chu.Unlock()

	chu.Valid = false

	return
}

// Signal that the current data of the chunk is valid again.
//
// This doesn't need to be called, if setImage() has been called.
func (chu *chunk) revalidate() {
	chu.Lock()
	defer chu.Unlock()

	chu.PixelQueue = []pixelQueueElement{}
	chu.Downloading = false
	chu.Valid = true

	return
}

// Signals that the data for the chunk is being downloaded.
// While a chunk is downloading, all setPixel() calls will be queued.
// A valid chunk or a chunk that is already downloading will not be set, and it's returned whether or not it could be set.
//
// setImage() has to be used to signal the end of the download, and make the chunk valid (in sync with the game).
func (chu *chunk) signalDownload() bool {
	chu.Lock()
	defer chu.Unlock()

	if chu.Valid || chu.Downloading {
		return false
	}

	chu.PixelQueue = []pixelQueueElement{} // Empty queue on new download.
	chu.Downloading = true

	return true
}

type chunkQueryResult int

const (
	chunkKeep chunkQueryResult = iota
	chunkDownload
	chunkCompress // TODO: Compress chunk after some time
	chunkDelete
)

// Query a chunk and reset its timer.
// The result suggests whether a chunk should be downloaded, kept or deleted.
// The canvas handles the result.
func (chu *chunk) getQueryState(resetTime bool) chunkQueryResult {
	chu.Lock()
	defer chu.Unlock()

	if chu.LastQueryTime.Add(chunkDeleteTime).Before(time.Now()) {
		return chunkDelete
	}
	if !chu.Valid && !chu.Downloading {
		return chunkDownload
	}

	// Only set the time when the chunk is not downloading. So it will be deleted after some time if it is "stuck"
	if !chu.Downloading && resetTime {
		chu.LastQueryTime = time.Now()
	}

	return chunkKeep
}
