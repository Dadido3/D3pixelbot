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

	"github.com/Dadido3/go-sciter"
	gorice "github.com/Dadido3/go-sciter/rice"
	"github.com/Dadido3/go-sciter/window"
)

// ONLY CALL FROM MAIN THREAD!
func sciterOpenMain() {
	//sciter.SetOption(sciter.SCITER_SET_DEBUG_MODE, 1)
	sciter.SetOption(sciter.SCITER_SET_SCRIPT_RUNTIME_FEATURES, sciter.ALLOW_FILE_IO|sciter.ALLOW_SOCKET_IO|sciter.ALLOW_EVAL|sciter.ALLOW_SYSINFO) // Needed for the inspector to work!

	w, err := window.New(sciter.SW_MAIN|sciter.SW_RESIZEABLE|sciter.SW_TITLEBAR|sciter.SW_CONTROLS|sciter.SW_ENABLE_DEBUG|sciter.SW_GLASSY, sciter.NewRect(300, 300, 500, 400)) // TODO: Store/Restore window position or open it in screen center
	if err != nil {
		log.Panic(err)
	}

	gorice.HandleDataLoad(w.Sciter)

	w.DefineFunction("openLocal", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			log.Errorf("Wrong number of parameters")
			return sciter.NewValue("Wrong number of parameters")
		}
		if !args[0].IsString() {
			log.Errorf("Wrong type of parameters")
			return sciter.NewValue("Wrong type of parameters")
		}

		game := args[0].String() // Always clone, otherwise those are just references to sciter values and will be invalid if used after return

		connectionType, ok := connectionTypes[game]
		if !ok {
			log.Errorf("game %v not found", game)
			return sciter.NewValue(fmt.Sprintf("game %v not found", game))
		}

		con, can := connectionType.FunctionNew()

		closeSignal := sciterOpenCanvas(con, can)

		go func() {
			<-closeSignal
			con.Close()
		}()

		return nil
	})

	w.DefineFunction("recordLocal", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			log.Errorf("Wrong number of parameters")
			return sciter.NewValue("Wrong number of parameters")
		}
		if !args[0].IsString() {
			log.Errorf("Wrong type of parameters")
			return sciter.NewValue("Wrong type of parameters")
		}

		game := args[0].String() // Always clone, otherwise those are just references to sciter values and will be invalid if used after return

		connectionType, ok := connectionTypes[game]
		if !ok {
			log.Errorf("game %v not found", game)
			return sciter.NewValue(fmt.Sprintf("game %v not found", game))
		}

		con, can := connectionType.FunctionNew()

		closeSignal := sciterOpenRecorder(con, can)

		go func() {
			<-closeSignal
			con.Close()
		}()

		return nil
	})

	w.DefineFunction("replayLocal", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			log.Errorf("Wrong number of parameters")
			return sciter.NewValue("Wrong number of parameters")
		}
		if !args[0].IsString() {
			log.Errorf("Wrong type of parameters")
			return sciter.NewValue("Wrong type of parameters")
		}

		game := args[0].String() // Always clone, otherwise those are just references to sciter values and will be invalid if used after return

		con, can, err := newCanvasDiskReader(game)
		if err != nil {
			log.Errorf("Can't open recording of %v: %v", game, err)
			return sciter.NewValue(fmt.Sprintf("Can't open recording of %v: %v", game, err))
		}

		closeSignal := sciterOpenCanvas(con, can)

		go func() {
			<-closeSignal
			con.Close()
		}()

		return nil
	})

	w.DefineFunction("version", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 0 {
			log.Errorf("Wrong number of parameters")
			return sciter.NewValue("Wrong number of parameters")
		}

		return sciter.NewValue(version.String())
	})

	if err := w.LoadFile("rice://ui/main.htm"); err != nil {
		log.Panic(err)
	}

	w.Show()
	w.Run()
}
