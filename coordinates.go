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

type pixelSize image.Point // Size of something in pixels

type (
	chunkCoordinate image.Point               // Coordinate of something in chunks
	chunkRectangle  struct{ image.Rectangle } // A rectangle something in chunks
	chunkSize       image.Point               // A size of something in chunks
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

	return chunkRectangle{image.Rectangle{
		Min: min,
		Max: max,
	}}
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

	return chunkRectangle{image.Rectangle{
		Min: min,
		Max: max,
	}}
}
