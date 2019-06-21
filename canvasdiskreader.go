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
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "golang.org/x/image/bmp"

	gzip "github.com/klauspost/pgzip"
)

type canvasDiskReader struct {
	ShortName string

	ChunkSize   pixelSize
	ChunkOrigin image.Point
	Canvas      *canvas
	Recordings  []canvasDiskReaderRecording

	TimeChan      chan time.Time // Sends point in time to goroutine
	QuitWaitGroup sync.WaitGroup
}

type canvasDiskReaderRecording struct {
	FileName           string
	StartTime, EndTime time.Time
}

func newCanvasDiskReader(shortName string) (connection, *canvas, error) {
	cdr := &canvasDiskReader{
		ShortName: shortName,
		TimeChan:  make(chan time.Time, 1),
	}

	var err error
	cdr.Recordings, err = cdr.refreshRecordings()
	if err != nil {
		return nil, nil, fmt.Errorf("Can't get recordings from %v", shortName)
	}

	if len(cdr.Recordings) <= 0 {
		return nil, nil, fmt.Errorf("Found no recordings for %v", shortName)
	}

	cdr.TimeChan <- cdr.Recordings[0].StartTime

	cdr.Canvas, _ = newCanvas(cdr.ChunkSize, cdr.ChunkOrigin, image.Rect(math.MinInt32, math.MinInt32, math.MaxInt32, math.MaxInt32))

	cdr.QuitWaitGroup.Add(1)
	go func() {
		defer cdr.QuitWaitGroup.Done()
		ticker := time.NewTicker(100 * time.Millisecond) // Ticker for sending time update events to the canvas
		defer ticker.Stop()

		defer log.Tracef("Closed replay goroutine of %v", shortName)

		destTime, ok := <-cdr.TimeChan // Destination time and channel state
		var replayTime time.Time

		// Run while channel is open
		for ok {
			// Get recording file where current time is inside its time interval
			var rec canvasDiskReaderRecording
			found := false
			for _, recording := range cdr.Recordings {
				if !destTime.Before(recording.StartTime) && destTime.Before(recording.EndTime) {
					rec = recording
					found = true
					break
				}
			}

			if !found {
				cdr.Canvas.setTime(destTime)
				destTime, ok = <-cdr.TimeChan
				continue
			}

			// Blocks while destTime < newReplayTime
			// Returns false when a (new) recording should be (re)opened
			waitTime := func(newReplayTime time.Time) bool {
				// Get next point in time
				select {
				case destTime, ok = <-cdr.TimeChan:
					if !ok {
						return false // Close goroutine
					}
					// Check if destination time is outside of the recording's time range
					if destTime.Before(rec.StartTime) || !destTime.Before(rec.EndTime) {
						return false
					}
					// Check if destination time is before replayTime
					if destTime.Before(replayTime) {
						return false
					}
				default:
				}

				// Block as long as destTime is < newReplayTime
				for destTime.Before(newReplayTime) {
					cdr.Canvas.setTime(destTime) // Output current time when waiting

					destTime, ok = <-cdr.TimeChan
					if !ok {
						return false // Close goroutine
					}
					// Check if destination time is outside of the recording's time range
					if destTime.Before(rec.StartTime) || !destTime.Before(rec.EndTime) {
						return false
					}
					// Check if destination time is before replayTime
					if destTime.Before(replayTime) {
						return false
					}
				}

				replayTime = newReplayTime
				select {
				case <-ticker.C:
					cdr.Canvas.setTime(replayTime) // Send out time update every xxx ms
				default:
				}

				return true
			}

			// Open and read recording. In a function, so defer works inside the loop
			func() {
				// Invalidate all on file close
				defer cdr.Canvas.invalidateAll()

				// Found valid recording, read it
				fileName := rec.FileName
				log.Debugf("Open recording %v", fileName)
				file, err := os.Open(fileName)
				if err != nil {
					log.Warnf("Can't open file %v: %v", fileName, err)
					waitTime(rec.EndTime)
					return
				}
				defer file.Close()
				zipReader, err := gzip.NewReader(file)
				if err != nil {
					log.Warnf("Can't decompress %v: %v", fileName, err)
					waitTime(rec.EndTime)
					return
				}
				defer zipReader.Close()

				var chunkSize pixelSize
				var chunkOrigin image.Point
				replayTime, chunkSize, chunkOrigin, err = canvasDiskReaderParseHeader(zipReader)
				if err != nil {
					log.Warn(err)
					waitTime(rec.EndTime)
					return
				}
				if cdr.Canvas.ChunkSize != chunkSize {
					log.Warnf("Chunk size differs in recording %v. From %v to %v. Seperate this and similar files from the others to play it", fileName, cdr.Canvas.ChunkSize, chunkSize)
					waitTime(rec.EndTime)
					return
				}
				if cdr.Canvas.Origin != chunkOrigin {
					log.Warnf("Origin differs in recording %v. From %v to %v. Seperate this and similar files from the others to play it", fileName, cdr.Canvas.Origin, chunkOrigin)
					waitTime(rec.EndTime)
					return
				}

				// Loop that retrieves all the events until replayTime >= destTime
				for {
					// Read and send events
					var dataType uint8
					var binTime int64
					err := binary.Read(zipReader, binary.LittleEndian, &dataType)
					if err != nil {
						log.Warnf("Error while reading file %v: %v", fileName, err)
						waitTime(rec.EndTime)
						return
					}
					err = binary.Read(zipReader, binary.LittleEndian, &binTime)
					if err != nil {
						log.Warnf("Error while reading file %v: %v", fileName, err)
						waitTime(rec.EndTime)
						return
					}

					// Block until time is progressed enough. Or if another file needs to be loaded (on false)
					if !waitTime(time.Unix(0, binTime)) {
						return
					}

					switch dataType {
					case 10: // SetPixel
						var dat struct {
							X, Y    int32
							R, G, B uint8
						}
						err := binary.Read(zipReader, binary.LittleEndian, &dat)
						if err != nil {
							log.Warnf("Error while reading file %v: %v", fileName, err)
							waitTime(rec.EndTime)
							return
						}
						cdr.Canvas.setPixel(image.Point{int(dat.X), int(dat.Y)}, color.RGBA{dat.R, dat.G, dat.B, 255})

					case 20: // InvalidateRect
						var dat struct {
							MinX, MinY, MaxX, MaxY int32
						}
						err := binary.Read(zipReader, binary.LittleEndian, &dat)
						if err != nil {
							log.Warnf("Error while reading file %v: %v", fileName, err)
							waitTime(rec.EndTime)
							return
						}
						cdr.Canvas.invalidateRect(image.Rect(int(dat.MinX), int(dat.MinY), int(dat.MaxX), int(dat.MaxY)))

					case 21: // InvalidateAll
						cdr.Canvas.invalidateAll()

					case 22: // RevalidateRect
						var dat struct {
							MinX, MinY, MaxX, MaxY int32
						}
						err := binary.Read(zipReader, binary.LittleEndian, &dat)
						if err != nil {
							log.Warnf("Error while reading file %v: %v", fileName, err)
							waitTime(rec.EndTime)
							return
						}
						cdr.Canvas.revalidateRect(image.Rect(int(dat.MinX), int(dat.MinY), int(dat.MaxX), int(dat.MaxY)))

					case 30: // SetImage
						var dat struct {
							X, Y int32
							Size uint32
						}
						err := binary.Read(zipReader, binary.LittleEndian, &dat)
						if err != nil {
							log.Warnf("Error while reading file %v: %v", fileName, err)
							waitTime(rec.EndTime)
							return
						}
						rawBytes := make([]byte, dat.Size)
						_, err = io.ReadFull(zipReader, rawBytes)
						if err != nil {
							log.Warnf("Error while reading file %v: %v", fileName, err)
							waitTime(rec.EndTime)
							return
						}
						img, imageFormat, err := image.Decode(bytes.NewBuffer(rawBytes))
						if err != nil {
							log.Warnf("Error while reading %v image from %v: %v", imageFormat, fileName, err)
							waitTime(rec.EndTime)
							return
						}

						// Move image to X and Y
						switch img := img.(type) {
						case *image.Paletted:
							img.Rect = img.Rect.Add(image.Point{int(dat.X), int(dat.Y)})
						case *image.RGBA:
							img.Rect = img.Rect.Add(image.Point{int(dat.X), int(dat.Y)})
						default:
							log.Warnf("Unknown internal image type %T in %v", img, fileName)
						}

						cdr.Canvas.signalDownload(img.Bounds())
						cdr.Canvas.setImage(img, false, true)

					default:
						log.Warnf("Found invalid data type %v in %v", dataType, fileName)
						waitTime(rec.EndTime)
						return

					}
				}
			}()
		}
	}()

	return cdr, cdr.Canvas, nil
}

func canvasDiskReaderParseHeader(reader io.Reader) (time.Time, pixelSize, image.Point, error) {
	var dat struct {
		MagicNumber             [4]byte
		Version                 uint16 // File format version
		Time                    int64
		ChunkWidth, ChunkHeight uint32
		OriginX, OriginY        int32  // Origin/Offset of the chunks
		_                       uint32 // Reserved // TODO: Somehow store endTime here
		_                       uint32 // Reserved
		_                       uint32 // Reserved
		_                       uint32 // Reserved
		_                       uint32 // Reserved
		_                       uint32 // Reserved
	}
	err := binary.Read(reader, binary.LittleEndian, &dat)
	if err != nil {
		return time.Time{}, pixelSize{}, image.Point{}, fmt.Errorf("Error while reading file: %v", err)
	}

	if dat.MagicNumber != [4]byte{'P', 'R', 'E', 'C'} {
		return time.Time{}, pixelSize{}, image.Point{}, fmt.Errorf("Wrong file format")
	}

	if dat.Version > 1 {
		return time.Time{}, pixelSize{}, image.Point{}, fmt.Errorf("Version is newer")
	}

	return time.Unix(0, dat.Time), pixelSize{int(dat.ChunkWidth), int(dat.ChunkHeight)}, image.Point{int(dat.OriginX), int(dat.OriginY)}, nil
}

func (cdr *canvasDiskReader) setReplayTime(t time.Time) error {
	// Write into channel, or replace the current element if the channel is full
	select {
	case cdr.TimeChan <- t:
	default:
		select {
		case <-cdr.TimeChan:
		default:
		}
		cdr.TimeChan <- t
	}

	return nil
}

// Creates list of recordings
func (cdr *canvasDiskReader) refreshRecordings() ([]canvasDiskReaderRecording, error) {
	fileDirectory := filepath.Join(wd, "recordings", cdr.ShortName)
	files, err := ioutil.ReadDir(fileDirectory)
	if err != nil {
		return nil, fmt.Errorf("Can't read from %v", fileDirectory)
	}

	// Filter pixrec files
	tempFiles := []os.FileInfo{}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".pixrec" {
			tempFiles = append(tempFiles, f)
		}
	}
	files = tempFiles

	recs := []canvasDiskReaderRecording{}

	// Get info of all recordings
	for i, file := range files {
		fileName := filepath.Join(fileDirectory, file.Name())
		f, err := os.Open(fileName)
		if err != nil {
			log.Warnf("Can't open recording %v", fileName)
			continue
		}
		defer f.Close()

		zipReader, err := gzip.NewReader(f)
		if err != nil {
			log.Warnf("Can't initialize gzip reader for %v: %v", fileName, err)
			continue
		}
		defer zipReader.Close()

		startTime, chunkSize, chunkOrigin, err := canvasDiskReaderParseHeader(zipReader)
		if err != nil {
			log.Warnf("Error reading header of %v: %v", fileName, err)
			continue
		}

		// Check if it fits to the stored chunk size and chunk origin
		empty := pixelSize{}
		if cdr.ChunkSize == empty {
			cdr.ChunkSize, cdr.ChunkOrigin = chunkSize, chunkOrigin
		}
		if cdr.ChunkSize != chunkSize {
			log.Warnf("Chunk size differs in recording %v. From %v to %v. Separate this and similar files from the others to play it", fileName, cdr.ChunkSize, chunkSize)
			continue
		}
		if cdr.ChunkOrigin != chunkOrigin {
			log.Warnf("Origin differs in recording %v. From %v to %v. Separate this and similar files from the others to play it", fileName, cdr.ChunkOrigin, chunkOrigin)
			continue
		}

		rec := canvasDiskReaderRecording{
			FileName:  fileName,
			StartTime: startTime,
			EndTime:   time.Now(), // Set it to "now", it will be overwritten by the next recording, if there is one
		}

		// Set the end time of the previous element to the start time of the current
		if i > 0 {
			recs[i-1].EndTime = startTime
		}

		recs = append(recs, rec)
	}

	return recs, nil
}

func (cdr *canvasDiskReader) getRecordings() []canvasDiskReaderRecording {
	return cdr.Recordings
}

func (cdr *canvasDiskReader) getShortName() string {
	return fmt.Sprintf("replay-%v", cdr.ShortName)
}

func (cdr *canvasDiskReader) getName() string {
	return fmt.Sprintf("Replay of %v", cdr.ShortName)
}

func (cdr *canvasDiskReader) getOnlinePlayers() int {
	return 0
}

// Closes the reader and the canvas
func (cdr *canvasDiskReader) Close() {
	// Stop goroutines gracefully
	close(cdr.TimeChan)
	cdr.QuitWaitGroup.Wait()

	cdr.Canvas.Close()

	return
}
