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

// TODO: Prevent a writer to be created for the same connection several times

package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	gzip "github.com/klauspost/pgzip"
)

type canvasDiskWriter struct {
	Closed      bool
	ClosedMutex sync.RWMutex

	Canvas *canvas

	File      *os.File
	ZipWriter *gzip.Writer
}

func (can *canvas) newCanvasDiskWriter(shortName string) (*canvasDiskWriter, error) {
	cdw := &canvasDiskWriter{
		Canvas: can,
	}

	re := regexp.MustCompile("[^a-zA-Z0-9\\-\\.]+")
	shortName = re.ReplaceAllString(shortName, "_")

	fileName := time.Now().UTC().Format("2006-01-02T150405") + ".pixrec" // Use RFC3339 like encoding, but with : removed
	fileDirectory := filepath.Join(".", "recordings", shortName)
	filePath := filepath.Join(fileDirectory, fileName)

	os.MkdirAll(fileDirectory, 0777)
	f, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("Can't create file %v: %v", filePath, err)
	}

	cdw.File = f
	zipWriter, err := gzip.NewWriterLevel(f, gzip.DefaultCompression)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("Can't initialize compression %v: %v", filePath, err)
	}
	cdw.ZipWriter = zipWriter

	// Write basic information about the canvas
	cdw.ZipWriter.Name = shortName
	cdw.ZipWriter.Comment = "D3's custom pixel game client recording"

	err = binary.Write(cdw.ZipWriter, binary.LittleEndian, struct {
		MagicNumber             uint32
		Version                 uint16 // File format version
		Time                    int64
		ChunkWidth, ChunkHeight uint32
		Reserved                uint16
	}{
		MagicNumber: 1128616528, // ASCII "PREC" in little endian
		Version:     1,
		Time:        time.Now().UnixNano(),
		ChunkWidth:  uint32(can.ChunkSize.X),
		ChunkHeight: uint32(can.ChunkSize.Y),
	})
	if err != nil {
		zipWriter.Close()
		f.Close()
		return nil, fmt.Errorf("Can't write to file %v: %v", filePath, err)
	}

	can.subscribeListener(cdw)

	return cdw, nil
}

func (cdw *canvasDiskWriter) setListeningRects(rects []image.Rectangle) error {
	cdw.ClosedMutex.RLock()
	defer cdw.ClosedMutex.RUnlock()
	if cdw.Closed {
		return fmt.Errorf("Listener is closed")
	}

	cdw.Canvas.registerRects(cdw, rects, true)

	return nil
}

func (cdw *canvasDiskWriter) handleSetPixel(pos image.Point, color color.Color) error {
	cdw.ClosedMutex.RLock()
	defer cdw.ClosedMutex.RUnlock()
	if cdw.Closed {
		return fmt.Errorf("Listener is closed")
	}

	r, g, b, _ := color.RGBA() // Returns 16 bit per channel

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
		R:        uint8(r >> 8),
		G:        uint8(g >> 8),
		B:        uint8(b >> 8),
	})
	if err != nil {
		return fmt.Errorf("Can't write to file %v: %v", cdw.File.Name(), err)
	}

	return nil
}

func (cdw *canvasDiskWriter) handleInvalidateRect(rect image.Rectangle) error {
	cdw.ClosedMutex.RLock()
	defer cdw.ClosedMutex.RUnlock()
	if cdw.Closed {
		return fmt.Errorf("Listener is closed")
	}

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
	cdw.ClosedMutex.RLock()
	defer cdw.ClosedMutex.RUnlock()
	if cdw.Closed {
		return fmt.Errorf("Listener is closed")
	}

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

func (cdw *canvasDiskWriter) handleRevalidateRect(rect image.Rectangle) error {
	cdw.ClosedMutex.RLock()
	defer cdw.ClosedMutex.RUnlock()
	if cdw.Closed {
		return fmt.Errorf("Listener is closed")
	}

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

func (cdw *canvasDiskWriter) handleSignalDownload(rect image.Rectangle) error {
	cdw.ClosedMutex.RLock()
	defer cdw.ClosedMutex.RUnlock()
	if cdw.Closed {
		return fmt.Errorf("Listener is closed")
	}

	// There is no need to write that data to disk
	// The signalDownload event will be simulated by the diskreader later

	return nil
}

func (cdw *canvasDiskWriter) handleSetImage(img image.Image, valid bool) error {
	cdw.ClosedMutex.RLock()
	defer cdw.ClosedMutex.RUnlock()
	if cdw.Closed {
		return fmt.Errorf("Listener is closed")
	}

	// If image is not in sync with the game, ignore it. A valid image will come later
	if !valid {
		return nil
	}

	bounds := img.Bounds()
	arrayRGB := imageToRGBArray(img) // TODO: Also write paletted image if there is one

	err := binary.Write(cdw.ZipWriter, binary.LittleEndian, struct {
		DataType      uint8
		Time          int64
		X, Y          int32
		Width, Height uint16
		Size          uint32 // Size of the RGB data in bytes
	}{
		DataType: 30,
		Time:     time.Now().UnixNano(),
		X:        int32(bounds.Min.X),
		Y:        int32(bounds.Min.Y),
		Width:    uint16(bounds.Dx()),
		Height:   uint16(bounds.Dy()),
		Size:     uint32(len(arrayRGB)),
	})
	if err != nil {
		return fmt.Errorf("Can't write to file %v: %v", cdw.File.Name(), err)
	}
	err = binary.Write(cdw.ZipWriter, binary.LittleEndian, arrayRGB)
	if err != nil {
		return fmt.Errorf("Can't write to file %v: %v", cdw.File.Name(), err)
	}
	return nil
}

func (cdw *canvasDiskWriter) handleChunksChange(create, remove []image.Rectangle) error {
	cdw.ClosedMutex.RLock()
	defer cdw.ClosedMutex.RUnlock()
	if cdw.Closed {
		return fmt.Errorf("Listener is closed")
	}

	// There is no need to write that data to disk

	return nil
}

func (cdw *canvasDiskWriter) Close() {
	cdw.Canvas.unsubscribeListener(cdw)
	cdw.handleInvalidateAll()

	cdw.ClosedMutex.RLock()
	cdw.Closed = true // Prevent any new events from happening
	cdw.ClosedMutex.RUnlock()

	cdw.ZipWriter.Close()
	cdw.File.Close()
}
