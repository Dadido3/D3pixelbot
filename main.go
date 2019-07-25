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
// TODO: Refactor most variable names when gorename works with modules
// TODO: Add headless mode, and service

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Dadido3/configdb"
	"github.com/coreos/go-semver/semver"
	colorable "github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var wd string // Initial working directory (Executable directory)
var version *semver.Version
var conf *configdb.Config

func init() {
	var err error
	wd, err = os.Getwd()
	if err != nil {
		log.Fatalf("Can't get working directory")
	}

	version, err = semver.NewVersion("0.1.4")
	if err != nil {
		fmt.Println(err.Error())
	}
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

	os.MkdirAll(filepath.Join(wd, "log"), os.ModePerm)
	f, err := os.OpenFile(filepath.Join(wd, "log", time.Now().UTC().Format("2006-01-02T150405")+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(io.MultiWriter(colorable.NewColorableStdout(), f)) // TODO: Separate formatting for logfiles
	log.SetLevel(logrus.TraceLevel)

	storages := []configdb.Storage{configdb.UseJSONFile("config.json")}
	conf, err = configdb.New(storages)
	if err != nil {
		log.Errorf("Can't load configuration: %v", err)
	}

	log.Infof("D3pixelbot %v started", version)

	/*pFile, err := os.Create("cpu.pprof")
	if err != nil {
		log.Fatal(err)
	}
	defer pFile.Close()
	pprof.StartCPUProfile(pFile)
	defer pprof.StopCPUProfile()*/

	sciterOpenMain()
}
