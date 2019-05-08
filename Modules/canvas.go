package main

import "image/color"

type pixelCoordinate struct {
	X, Y int
}
type pixelSize pixelCoordinate

type canvas struct {
	Chunks map[chunkCoordinate]*chunk

	ChunkSize pixelSize
	Palette   color.Palette
}

func newCanvas(chunkSize pixelSize, palette color.Palette) (*canvas, error) {
	can := &canvas{
		Chunks:    make(map[chunkCoordinate]*chunk, 0),
		ChunkSize: chunkSize,
		Palette:   palette,
	}

	return can, nil
}

func (can *canvas) pixelToChunkCoord(coord pixelCoordinate) chunkCoordinate {
	return chunkCoordinate{
		X: divideFloor(coord.X, can.ChunkSize.X),
		Y: divideFloor(coord.Y, can.ChunkSize.Y),
	}
}
