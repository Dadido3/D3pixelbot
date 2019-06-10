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

// TODO: Change channels to be handled and closed by the sending side, to prevent write access to already closed channels.
// TODO: Redo most of the goroutine stopping mechanism
// TODO: Add manifest for DPI awareness: https://github.com/c-smile/sciter-sdk/blob/master/demos/usciter/win-res/dpi-aware.manifest
// TODO: Add way to gracefully stop everything when main window closes, or when the console closes.

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	colorable "github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var log = logrus.New()

func init() {
	runtime.LockOSThread() // Locks the whole program to the main thread (Except newly spawned goroutines). That's needed for the UI to work properly.
	// TODO: Only lock when UI is needed (If headless mode is configured), or use lib to call sciterOpenMain() from main thread.
}

func main() {
	log.SetReportCaller(true)
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			//return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
			return fmt.Sprintf("%s()", f.Function), ""
		},
	})

	os.MkdirAll(filepath.Join(".", "log"), os.ModePerm)
	f, err := os.OpenFile(filepath.Join(".", "log", time.Now().UTC().Format("2006-01-02T150405")+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(io.MultiWriter(colorable.NewColorableStdout(), f)) // TODO: Separate formatting for logfiles
	log.SetLevel(logrus.TraceLevel)

	viper.SetConfigFile(filepath.Join(".", "config.json"))
	err = viper.ReadInConfig()
	if err != nil {
		log.Errorf("Can't load config file: %v", err)
	}

	log.Info("D3pixelbot started")

	/*pFile, err := os.Create("profile.out")
	if err != nil {
		log.Fatal(err)
	}
	defer pFile.Close()
	pprof.StartCPUProfile(pFile)
	defer pprof.StopCPUProfile()*/

	sciterOpenMain()
}
