package main

// TODO: Write changes to disk

import (
	"fmt"
	"image"
	"image/color"
	"sync"
)

type canvas struct {
	sync.RWMutex

	Chunks map[chunkCoordinate]*chunk

	ChunkSize pixelSize
	Palette   color.Palette
}

func newCanvas(chunkSize pixelSize, palette color.Palette) *canvas {
	can := &canvas{
		Chunks:    make(map[chunkCoordinate]*chunk, 0),
		ChunkSize: chunkSize,
		Palette:   palette,
	}

	return can
}

func (can *canvas) getChunk(coord chunkCoordinate, createIfNonexistent bool) (*chunk, error) {
	can.RLock()
	defer can.RUnlock()

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

func (can *canvas) getChunks(rect chunkRectangle, createIfNonexistent bool) ([]*chunk, error) {
	rectTemp := image.Rectangle(rect).Canon()
	chunks := []*chunk{}

	for iy := rectTemp.Min.Y; iy < rectTemp.Max.Y; iy++ {
		for ix := rectTemp.Min.X; ix < rectTemp.Max.X; ix++ {
			chunk, err := can.getChunk(chunkCoordinate{ix, iy}, createIfNonexistent)
			if err != nil {
				// This assumes that there can only be an error when createIfNonexistent == false
				// So it will never abort while it creates chunks
				return nil, fmt.Errorf("Can't get all chunks: %v", err)
			}
			chunks = append(chunks, chunk)
		}
	}

	return chunks, nil
}
