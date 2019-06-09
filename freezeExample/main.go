package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"
)

func openWindowMain() {
	w, err := window.New(sciter.SW_MAIN|sciter.SW_RESIZEABLE|sciter.SW_TITLEBAR|sciter.SW_CONTROLS|sciter.SW_GLASSY|sciter.SW_ENABLE_DEBUG, sciter.NewRect(100, 100, 550, 570))
	if err != nil {
		log.Fatal("sciter create window failed", err)
	}

	w.DefineFunction("openLocal", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			return sciter.NewValue("Wrong number of parameters")
		}
		if !args[0].IsString() {
			return sciter.NewValue("Wrong type of parameters")
		}

		args[0].String()

		openWindowCanvas()

		return nil
	})

	path, err := filepath.Abs("ui/main.htm")
	if err != nil {
		log.Fatal(err)
	}

	if err := w.LoadFile("file://" + path); err != nil {
		log.Fatal(err)
	}

	w.Show()
	w.Run()
}

func openWindowCanvas() {
	w, err := window.New(sciter.SW_RESIZEABLE|sciter.SW_TITLEBAR|sciter.SW_CONTROLS|sciter.SW_GLASSY|sciter.SW_ENABLE_DEBUG, sciter.NewRect(300, 300, 550, 570))
	if err != nil {
		log.Fatal("sciter create window failed", err)
	}

	path, err := filepath.Abs("ui/canvas.htm")
	if err != nil {
		log.Fatal(err)
	}

	if err := w.LoadFile("file://" + path); err != nil {
		log.Fatal(err)
	}

	w.Show()
}

func main() {
	sciter.SetOption(sciter.SCITER_SET_SCRIPT_RUNTIME_FEATURES, sciter.ALLOW_FILE_IO|sciter.ALLOW_SOCKET_IO|sciter.ALLOW_EVAL|sciter.ALLOW_SYSINFO) // Needed for the inspector to work!

	for i := 0; i < 500; i++ {
		go func() {
			for {
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	openWindowMain()
}
