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
