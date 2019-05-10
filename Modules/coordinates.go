package main

import "image"

type pixelSize image.Point // Size of something in pixels

type (
	chunkCoordinate image.Point     // Coordinate of something in chunks
	chunkRectangle  image.Rectangle // A rectangle something in chunks
	chunkSize       image.Point     // A size of something in chunks
)

// Converts a pixel coordinate into a chunk coordinate containing the given pixel
func (ps pixelSize) getChunkCoord(coord image.Point) chunkCoordinate {
	return chunkCoordinate{
		X: divideFloor(coord.X, ps.X),
		Y: divideFloor(coord.Y, ps.Y),
	}
}

// Converts a pixel rectangle into the closest possible rectangle in chunk coordinates.
// The given pixel rectangle will always be inside or equal to the resulting chunk rectangle.
func (ps pixelSize) getOuterChunkRect(rect image.Rectangle) chunkRectangle {
	rectTemp := rect.Canon()

	min := image.Point{
		X: divideFloor(rectTemp.Min.X, ps.X),
		Y: divideFloor(rectTemp.Min.Y, ps.Y),
	}
	max := image.Point{
		X: divideCeil(rectTemp.Max.X, ps.X),
		Y: divideCeil(rectTemp.Max.Y, ps.Y),
	}

	return chunkRectangle{
		Min: min,
		Max: max,
	}
}

// Converts a pixel rectangle into the closest possible rectangle in chunk coordinates.
// The resulting chunk rectangle will always be inside or equal to the given pixel rectangle.
//
// Be aware that the resulting rectangle can have a length of 0 in any axis!
func (ps pixelSize) getInnerChunkRect(rect image.Rectangle) chunkRectangle {
	rectTemp := rect.Canon()

	min := image.Point{
		X: divideCeil(rectTemp.Min.X, ps.X),
		Y: divideCeil(rectTemp.Min.Y, ps.Y),
	}
	max := image.Point{
		X: divideFloor(rectTemp.Max.X, ps.X),
		Y: divideFloor(rectTemp.Max.Y, ps.Y),
	}

	if max.X < min.X {
		max.X = min.X
	}
	if max.Y < min.Y {
		max.Y = min.Y
	}

	return chunkRectangle{
		Min: min,
		Max: max,
	}
}
