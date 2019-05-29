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
	"fmt"
	"log"
	"path/filepath"

	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"
)

func sciterOpenMain() {
	sciter.SetOption(sciter.SCITER_SET_DEBUG_MODE, 1)
	sciter.SetOption(sciter.SCITER_SET_SCRIPT_RUNTIME_FEATURES, sciter.ALLOW_FILE_IO|sciter.ALLOW_SOCKET_IO|sciter.ALLOW_EVAL|sciter.ALLOW_SYSINFO) // Needed for the inspector to work!

	w, err := window.New(sciter.SW_MAIN|sciter.SW_RESIZEABLE|sciter.SW_TITLEBAR|sciter.SW_CONTROLS|sciter.SW_ENABLE_DEBUG /*|sciter.SW_GLASSY*/, sciter.DefaultRect)
	if err != nil {
		log.Fatal(err)
	}

	path, err := filepath.Abs("ui/main.htm")
	if err != nil {
		log.Fatal(err)
	}

	if err := w.LoadFile("file://" + path); err != nil {
		log.Fatal(err)
	}

	ok := w.SetOption(sciter.SCITER_SET_DEBUG_MODE, 1)
	if !ok {
		log.Println("set debug mode failed")
	}

	w.DefineFunction("openLocal", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			return sciter.NewValue("Wrong number of parameters")
		}
		if !args[0].IsString() {
			return sciter.NewValue("Wrong type of parameters")
		}

		game := args[0].String() // Always clone, otherwise those are just references to sciter values and will be invalid if used after return

		connectionType, ok := connectionTypes[game]
		if !ok {
			return sciter.NewValue(fmt.Sprintf("game %v not found", game))
		}

		log.Println(game)

		con, can, err := connectionType.FunctionNew(true)
		if err != nil {
			return sciter.NewValue(fmt.Sprintf("Couldn't create connection: %v", err))
		}

		go func() {
			defer con.Close()

			sciterOpenCanvas(con, can)

		}()

		return nil
	})

	w.Show()
	w.Run()
}
