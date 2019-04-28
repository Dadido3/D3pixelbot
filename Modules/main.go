package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	_, err := newPixelcanvasio()
	if err != nil {
		log.Fatal(err)
	}

	buf := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	sentence, err := buf.ReadBytes('\n')
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(sentence))
	}
}
