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
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"net/http"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func postJSON(url string, origin string, structure interface{}) (statusCode int, headers http.Header, bodyString []byte, err error) {
	jsonStr, err := json.Marshal(structure)
	if err != nil {
		return 0, nil, nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", origin)

	resp, err := myClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, resp.Header, body, nil
}

// Integer division that rounds to the next integer towards negative infinity
func divideFloor(a, b int) int {
	temp := a / b

	if ((a ^ b) < 0) && (a%b != 0) {
		return temp - 1
	}

	return temp
}

// Integer division that rounds to the next integer towards positive infinity
func divideCeil(a, b int) int {
	temp := a / b

	if ((a ^ b) >= 0) && (a%b != 0) {
		return temp + 1
	}

	return temp
}

func isPaletteEqual(pal1, pal2 color.Palette) bool {
	if len(pal1) != len(pal2) {
		return false
	}

	for k, col := range pal1 {
		r1, g1, b1, a1 := col.RGBA()
		r2, g2, b2, a2 := pal2[k].RGBA()
		if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
			return false
		}
	}

	return true
}

// Creates a copy of an image
func copyImage(img image.Image) (image.Image, error) {
	switch img := img.(type) {
	case *image.RGBA:
		imgCopy := &image.RGBA{
			Pix:    make([]uint8, len(img.Pix)),
			Stride: img.Stride,
			Rect:   img.Rect,
		}
		copy(imgCopy.Pix, img.Pix)
		return imgCopy, nil

	case *image.Paletted:
		imgCopy := &image.Paletted{
			Pix:     make([]uint8, len(img.Pix)),
			Stride:  img.Stride,
			Rect:    img.Rect,
			Palette: make(color.Palette, len(img.Palette)),
		}
		copy(imgCopy.Pix, img.Pix)
		copy(imgCopy.Palette, img.Palette)
		return imgCopy, nil
	}

	return nil, fmt.Errorf("Incompatible image type %T", img)
}

// Converts any image to an RGBA array
func imageToRGBAArray(img image.Image) []byte {
	rect := img.Bounds()

	switch img := img.(type) {
	case *image.RGBA:
		// Assumes that the stride == width * 4
		return img.Pix
	default:
		array := make([]byte, rect.Dx()*rect.Dy()*4)

		i := 0
		for iy := rect.Min.Y; iy < rect.Max.Y; iy++ {
			for ix := rect.Min.X; ix < rect.Max.X; ix++ {
				r, g, b, a := img.At(ix, iy).RGBA()
				array[i] = byte(r)
				i++
				array[i] = byte(g)
				i++
				array[i] = byte(b)
				i++
				array[i] = byte(a)
				i++
			}
		}

		return array
	}

}

// Converts any image to an RGB array
func imageToRGBArray(img image.Image) []byte {
	rect := img.Bounds()

	switch img := img.(type) {
	default:
		array := make([]byte, rect.Dx()*rect.Dy()*3)

		i := 0
		for iy := rect.Min.Y; iy < rect.Max.Y; iy++ {
			for ix := rect.Min.X; ix < rect.Max.X; ix++ {
				r, g, b, _ := img.At(ix, iy).RGBA()
				array[i] = byte(r)
				i++
				array[i] = byte(g)
				i++
				array[i] = byte(b)
				i++
			}
		}

		return array
	}

}

// Converts any image to an BGRA array
func imageToBGRAArray(img image.Image) []byte {
	rect := img.Bounds()

	switch img := img.(type) {
	case *image.RGBA:
		// Assumes that the stride == width * 4
		array := make([]uint8, len(img.Pix))
		copy(array, img.Pix)
		for i := 0; i <= len(img.Pix)-4; i += 4 {
			array[i], array[i+2] = array[i+2], array[i]
		}
		return array
	default:
		array := make([]byte, rect.Dx()*rect.Dy()*4)

		i := 0
		for iy := rect.Min.Y; iy < rect.Max.Y; iy++ {
			for ix := rect.Min.X; ix < rect.Max.X; ix++ {
				r, g, b, a := img.At(ix, iy).RGBA()
				array[i] = byte(b)
				i++
				array[i] = byte(g)
				i++
				array[i] = byte(r)
				i++
				array[i] = byte(a)
				i++
			}
		}

		return array
	}

}
