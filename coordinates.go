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

import "image"

// TODO: Use better names than "chunkSize", ...

type pixelSize image.Point // Size of something measured in pixels

type (
	chunkCoordinate image.Point               // Coordinate of something measured in chunks
	chunkRectangle  struct{ image.Rectangle } // A rectangle something measured in chunks
	chunkSize       image.Point               // A size of something measured in chunks
)

// Converts a pixel coordinate into a chunk coordinate containing the given pixel
//
// Positive origin values move the chunk grid into negative direction
func (ps pixelSize) getChunkCoord(coord image.Point, origin image.Point) chunkCoordinate {
	return chunkCoordinate{
		X: divideFloor(coord.X+origin.X, ps.X),
		Y: divideFloor(coord.Y+origin.Y, ps.Y),
	}
}

// Converts a pixel rectangle into the closest possible rectangle in chunk coordinates.
// The given pixel rectangle will always be inside or equal to the resulting chunk rectangle.
//
// Positive origin values move the chunk grid into negative direction
func (ps pixelSize) getOuterChunkRect(rect image.Rectangle, origin image.Point) chunkRectangle {
	rectTemp := rect.Canon().Add(origin)

	min := image.Point{
		X: divideFloor(rectTemp.Min.X, ps.X),
		Y: divideFloor(rectTemp.Min.Y, ps.Y),
	}
	max := image.Point{
		X: divideCeil(rectTemp.Max.X, ps.X),
		Y: divideCeil(rectTemp.Max.Y, ps.Y),
	}

	return chunkRectangle{image.Rectangle{
		Min: min,
		Max: max,
	}}
}

// Converts a pixel rectangle into the closest possible rectangle in chunk coordinates.
// The resulting chunk rectangle will always be inside or equal to the given pixel rectangle.
//
// Be aware that the resulting rectangle can have a length of 0 in any axis!
//
// Positive origin values move the chunk grid into negative direction
func (ps pixelSize) getInnerChunkRect(rect image.Rectangle, origin image.Point) chunkRectangle {
	rectTemp := rect.Canon().Add(origin)

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

	return chunkRectangle{image.Rectangle{
		Min: min,
		Max: max,
	}}
}

// Converts a rectangle from chunk coordinates into pixel coordinates
//
// Positive origin values moves the chunk grid into negative direction
func (ps chunkRectangle) getPixelRectangle(chunkSize pixelSize, origin image.Point) image.Rectangle {
	rectTemp := ps.Canon()

	rectTemp.Min.X *= chunkSize.X
	rectTemp.Min.Y *= chunkSize.Y
	rectTemp.Max.X *= chunkSize.X
	rectTemp.Max.Y *= chunkSize.Y

	return image.Rectangle(rectTemp).Sub(origin)
}

// Converts a size in chunks into a size in pixels
func (cs chunkSize) getPixelSize(chunkSize pixelSize) pixelSize {
	return pixelSize{
		X: cs.X * chunkSize.X,
		Y: cs.Y * chunkSize.Y,
	}
}
