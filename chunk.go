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

const (
	chunkDeleteTime = 5 * time.Minute
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
	// TODO: Rewrite to handle any image type. So it can handle arbitrary colors from recordings

	PixelQueue         []pixelQueueElement // Queued pixels, that are set while the image is downloading
	Valid, Downloading bool                // Valid: Data is in sync with the game. Downloading: Data is being downloaded
	LastQueryTime      time.Time           // Point in time, when that chunk was queried last time. If this chunk hasn't been queried for some period, it will be unloaded.
}

func newChunk(rect image.Rectangle, p color.Palette) *chunk {
	chunk := &chunk{
		Rect:          rect.Canon(),
		Palette:       p,
		Image:         image.NewPaletted(rect, p),
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

	// TODO: Check if colorIndex is valid

	if chu.Valid {
		chu.Image.SetColorIndex(pos.X, pos.Y, colorIndex)
	}

	// If chunk is downloading, append to queue to draw them later
	if chu.Downloading {
		chu.PixelQueue = append(chu.PixelQueue, pixelQueueElement{
			Pos:        pos,
			ColorIndex: colorIndex,
		})
	}

	return nil
}

// Overwrites the image data, validates the chunk and resets the downloading flag.
// The chunk boundaries needs to be inside the image boundaries, otherwise the operation will fail.
//
// All queued pixels will be replayed when this function is called.
// This helps to prevent inconsistencies while downloading chunks.
// The result image is an up to date copy containing all queued changes.
func (chu *chunk) setImage(img image.Image) (*image.Paletted, error) {
	chu.Lock()
	defer chu.Unlock()

	if !chu.Rect.In(img.Bounds()) {
		return nil, fmt.Errorf("The image doesn't fill the chunk completely")
	}

	draw.Draw(chu.Image, chu.Image.Rect, img, chu.Image.Rect.Min, draw.Over)

	// Replay all the queued pixels
	for _, pqe := range chu.PixelQueue {
		chu.Image.SetColorIndex(pqe.Pos.X, pqe.Pos.Y, pqe.ColorIndex)
	}

	chu.PixelQueue = []pixelQueueElement{}
	chu.Valid = true
	chu.Downloading = false

	imgCopy := *chu.Image
	copy(imgCopy.Pix, chu.Image.Pix)
	copy(imgCopy.Palette, chu.Image.Palette)

	return &imgCopy, nil
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
//
// setImage() has to be used to signal that the chunk is valid (in sync with the game).
func (chu *chunk) invalidateImage() {
	chu.Lock()
	defer chu.Unlock()

	chu.Valid = false

	return
}

// Signals that the data for the chunk is being downloaded.
// While a chunk is downloading, all setPixel() calls will be queued.
//
// setImage() has to be used to signal the end of the download, and make the chunk valid (in sync with the game).
func (chu *chunk) signalDownload() {
	chu.Lock()
	defer chu.Unlock()

	chu.PixelQueue = []pixelQueueElement{} // Empty queue on new download.
	chu.Downloading = true

	return
}

type chunkQueryResult int

const (
	chunkKeep chunkQueryResult = iota
	chunkDownload
	chunkDelete
)

// Query a chunk and reset its timer.
// The result suggests whether a chunk should be downloaded, kept or deleted.
// The canvas handles the result.
func (chu *chunk) getQueryState() chunkQueryResult {
	chu.Lock()
	defer chu.Unlock()

	if chu.LastQueryTime.Add(chunkDeleteTime).After(time.Now()) {
		return chunkDelete
	}
	if !chu.Valid && !chu.Downloading {
		return chunkDownload
	}

	// Only set the time when the chunk is valid and not downloading. So it will be deleted after time if it is "stuck"
	if chu.Valid && !chu.Downloading {
		chu.LastQueryTime = time.Now()
	}

	return chunkKeep
}
