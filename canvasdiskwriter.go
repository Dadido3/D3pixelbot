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

// TODO: Prevent any more actions after Close() has been called

package main

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type canvasDiskWriter struct {
	Canvas     *canvas
	Rectangles []image.Rectangle      // A list of canvas areas, that should be kept up to date automatically
	RectsChan  chan []image.Rectangle // Moves the list of rectangles to the goroutine internally

	File      *os.File
	ZipWriter *gzip.Writer
}

func (can *canvas) newCanvasDiskWriter(name string) (*canvasDiskWriter, error) {
	cdw := &canvasDiskWriter{
		Canvas:    can,
		RectsChan: make(chan []image.Rectangle),
	}

	re := regexp.MustCompile("[^a-zA-Z0-9\\-\\.]+")
	name = re.ReplaceAllString(name, "_")

	fileName := time.Now().UTC().Format("2006-01-02T150405") + ".pixrec" // Use RFC3339 like encoding, but with : removed
	fileDirectory := filepath.Join(".", "Recordings", name)
	filePath := filepath.Join(fileDirectory, fileName)

	os.MkdirAll(fileDirectory, 0777)
	f, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("Can't create file %v: %v", filePath, err)
	}

	cdw.File = f
	zipWriter, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("Can't initialize compression %v: %v", filePath, err)
	}
	cdw.ZipWriter = zipWriter

	// Write basic information about the canvas
	cdw.ZipWriter.Name = name
	cdw.ZipWriter.Comment = "D3's custom pixel game client recording"

	err = binary.Write(cdw.ZipWriter, binary.LittleEndian, struct {
		MagicNumber             uint32
		Version                 uint16 // File format version
		Time                    int64
		ChunkWidth, ChunkHeight uint32
		PaletteSize             uint16 // Size in entries, max 256
	}{
		MagicNumber: 1128616528, // ASCII "PREC" in little endian
		Version:     1,
		Time:        time.Now().UnixNano(),
		ChunkWidth:  uint32(can.ChunkSize.X),
		ChunkHeight: uint32(can.ChunkSize.Y),
		PaletteSize: uint16(len(can.Palette)),
	})
	if err != nil {
		zipWriter.Close()
		f.Close()
		return nil, fmt.Errorf("Can't write to file %v: %v", filePath, err)
	}

	// Embed the palette. It's not used other than to initialize the canvas. (This will be removed when the canvas supports arbitrary colors)
	for _, color := range can.Palette {
		r, g, b, _ := color.RGBA()
		binary.Write(cdw.ZipWriter, binary.LittleEndian, uint8(r))
		binary.Write(cdw.ZipWriter, binary.LittleEndian, uint8(g))
		binary.Write(cdw.ZipWriter, binary.LittleEndian, uint8(b))
	}
	// TODO: Handle errors in the palette writer

	can.subscribeListener(cdw)

	// Goroutine that queries rectangles
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				for _, rect := range cdw.Rectangles {
					can.queryRect(rect)
				}
			case rects, ok := <-cdw.RectsChan:
				if !ok {
					// Close goroutine, as the channel is gone
					return
				}
				cdw.Rectangles = rects
				for _, rect := range cdw.Rectangles {
					can.queryRect(rect)
				}
			}
		}
	}()

	return cdw, nil
}

func (cdw *canvasDiskWriter) setListeningRects(rects []image.Rectangle) {
	cdw.RectsChan <- rects // TODO: Don't write to channel after Close() has been called
}

func (cdw *canvasDiskWriter) handleSetPixel(pos image.Point, colorIndex uint8) error {
	if int(colorIndex) >= len(cdw.Canvas.Palette) {
		return fmt.Errorf("Index outside of palette")
	}
	r, g, b, _ := cdw.Canvas.Palette[colorIndex].RGBA()

	err := binary.Write(cdw.ZipWriter, binary.LittleEndian, struct {
		DataType uint8
		Time     int64
		X, Y     int32
		R, G, B  uint8
	}{
		DataType: 10,
		Time:     time.Now().UnixNano(),
		X:        int32(pos.X),
		Y:        int32(pos.Y),
		R:        uint8(r),
		G:        uint8(g),
		B:        uint8(b),
	})
	if err != nil {
		return fmt.Errorf("Can't write to file %v: %v", cdw.File.Name(), err)
	}
	return nil
}

func (cdw *canvasDiskWriter) handleInvalidateRect(rect image.Rectangle) error {
	err := binary.Write(cdw.ZipWriter, binary.LittleEndian, struct {
		DataType               uint8
		Time                   int64
		MinX, MinY, MaxX, MaxY int32
	}{
		DataType: 20,
		Time:     time.Now().UnixNano(),
		MinX:     int32(rect.Min.X),
		MinY:     int32(rect.Min.Y),
		MaxX:     int32(rect.Max.X),
		MaxY:     int32(rect.Max.Y),
	})
	if err != nil {
		return fmt.Errorf("Can't write to file %v: %v", cdw.File.Name(), err)
	}
	return nil
}

func (cdw *canvasDiskWriter) handleInvalidateAll() error {
	err := binary.Write(cdw.ZipWriter, binary.LittleEndian, struct {
		DataType uint8
		Time     int64
	}{
		Time:     time.Now().UnixNano(),
		DataType: 21,
	})
	if err != nil {
		return fmt.Errorf("Can't write to file %v: %v", cdw.File.Name(), err)
	}
	return nil
}

func (cdw *canvasDiskWriter) handleSignalDownload(rect image.Rectangle) error {
	err := binary.Write(cdw.ZipWriter, binary.LittleEndian, struct {
		DataType               uint8
		Time                   int64
		MinX, MinY, MaxX, MaxY int32
	}{
		DataType: 22,
		Time:     time.Now().UnixNano(),
		MinX:     int32(rect.Min.X),
		MinY:     int32(rect.Min.Y),
		MaxX:     int32(rect.Max.X),
		MaxY:     int32(rect.Max.Y),
	})
	if err != nil {
		return fmt.Errorf("Can't write to file %v: %v", cdw.File.Name(), err)
	}
	return nil
}

func (cdw *canvasDiskWriter) handleSetImage(img *image.Paletted) error {
	bounds := img.Bounds()
	imgRGBA := image.NewRGBA(bounds)
	draw.Draw(imgRGBA, bounds, img, bounds.Min, draw.Over)
	arrayRGBA := imgRGBA.Pix

	err := binary.Write(cdw.ZipWriter, binary.LittleEndian, struct {
		DataType      uint8
		Time          int64
		X, Y          int32
		Width, Height uint16
		Size          uint32 // Size of the RGBA data in bytes // TODO: Reduce the image data to just RGB
	}{
		DataType: 30,
		Time:     time.Now().UnixNano(),
		X:        int32(bounds.Min.X),
		Y:        int32(bounds.Min.Y),
		Width:    uint16(bounds.Dx()),
		Height:   uint16(bounds.Dy()),
		Size:     uint32(len(arrayRGBA)),
	})
	if err != nil {
		return fmt.Errorf("Can't write to file %v: %v", cdw.File.Name(), err)
	}
	err = binary.Write(cdw.ZipWriter, binary.LittleEndian, arrayRGBA)
	if err != nil {
		return fmt.Errorf("Can't write to file %v: %v", cdw.File.Name(), err)
	}
	return nil
}

func (cdw *canvasDiskWriter) Close() {
	cdw.Canvas.unsubscribeListener(cdw)

	close(cdw.RectsChan) // This will stop the goroutine after all events are processed

	cdw.handleInvalidateAll()
	cdw.ZipWriter.Close()
	cdw.File.Close()
}
