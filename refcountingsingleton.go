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
	"sync"
)

// This makes sure that there is only a single shared instance of some structure, similar to a singleton.
// Additionally this will init, free and reinit the instance based on a reference counter.
type refCountingSingleton struct {
	sync.Mutex

	object  interface{}
	counter int
}

// Returns a pointer to a new or shared instance, and increases its reference counter.
// In case there is no object, `init` is called to create one.
func (s *refCountingSingleton) get(init func() interface{}) interface{} {
	s.Lock()
	defer s.Unlock()

	if s.object != nil {
		s.counter++
		return s.object
	}

	// There is no object, create new one
	s.object = init()
	s.counter = 1
	return s.object
}

// Decreases the reference counter of the object.
// If the reference counter reaches 0, true is returned to trigger some cleanup routine if necessary.
func (s *refCountingSingleton) release(object interface{}) bool {
	s.Lock()
	defer s.Unlock()

	if s.object != object {
		panic("Trying to release wrong object")
	}
	if s.object == nil {
		panic("Trying to release nil object")
	}

	s.counter--

	if s.counter == 0 {
		s.object = nil
		return true
	}

	return false
}
