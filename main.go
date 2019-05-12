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
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	//_, err := newPixelcanvasio()
	//if err != nil {
	//	log.Fatal(err)
	//}

	canvas := newCanvas(pixelSize{64, 64}, pixelcanvasioPalette)
	log.Println(canvas.getChunk(chunkCoordinate{0, 0}, false))

	buf := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	sentence, err := buf.ReadBytes('\n')
	if err != nil {
		log.Panicln(err)
	} else {
		fmt.Println(string(sentence))
	}
}
