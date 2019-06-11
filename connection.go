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

import "time"

type connection interface {
	getShortName() string // Return short and filesystem friendly name, also used as internal identifier
	getName() string      // Return full name for display purposes

	// TODO: Add subscribe and unsubscribe methods

	getOnlinePlayers() int
	Close()
}

// Same as connection, but it has some additional methods to set the replay time
type connectionReplay interface {
	connection

	setReplayTime(t time.Time) error
}

type connectionType struct {
	Name string

	FunctionNew func() (connection, *canvas)
}

var connectionTypes = map[string]connectionType{}
