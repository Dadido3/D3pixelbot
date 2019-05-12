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

import "testing"

func Test_divideFloor(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		struct{ a, b, want int }{-10, 1, -10},
		struct{ a, b, want int }{-1, 1, -1},
		struct{ a, b, want int }{0, 1, 0},
		struct{ a, b, want int }{1, 1, 1},
		struct{ a, b, want int }{10, 1, 10},
		struct{ a, b, want int }{-10, 2, -5},
		struct{ a, b, want int }{-1, 2, -1},
		struct{ a, b, want int }{0, 2, 0},
		struct{ a, b, want int }{1, 2, 0},
		struct{ a, b, want int }{10, 2, 5},
		struct{ a, b, want int }{-10, 3, -4},
		struct{ a, b, want int }{-1, 3, -1},
		struct{ a, b, want int }{0, 3, 0},
		struct{ a, b, want int }{1, 3, 0},
		struct{ a, b, want int }{10, 3, 3},
		struct{ a, b, want int }{-10, -1, 10},
		struct{ a, b, want int }{-1, -1, 1},
		struct{ a, b, want int }{0, -1, 0},
		struct{ a, b, want int }{1, -1, -1},
		struct{ a, b, want int }{10, -1, -10},
		struct{ a, b, want int }{-10, -2, 5},
		struct{ a, b, want int }{-1, -2, 0},
		struct{ a, b, want int }{0, -2, 0},
		struct{ a, b, want int }{1, -2, -1},
		struct{ a, b, want int }{10, -2, -5},
		struct{ a, b, want int }{-10, -3, 3},
		struct{ a, b, want int }{-1, -3, 0},
		struct{ a, b, want int }{0, -3, 0},
		struct{ a, b, want int }{1, -3, -1},
		struct{ a, b, want int }{10, -3, -4},
	}

	for _, test := range tests {
		if got := divideFloor(test.a, test.b); got != test.want {
			t.Errorf("divideFloor(%v, %v) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func Test_divideCeil(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		struct{ a, b, want int }{-10, 1, -10},
		struct{ a, b, want int }{-1, 1, -1},
		struct{ a, b, want int }{0, 1, 0},
		struct{ a, b, want int }{1, 1, 1},
		struct{ a, b, want int }{10, 1, 10},
		struct{ a, b, want int }{-10, 2, -5},
		struct{ a, b, want int }{-1, 2, 0},
		struct{ a, b, want int }{0, 2, 0},
		struct{ a, b, want int }{1, 2, 1},
		struct{ a, b, want int }{10, 2, 5},
		struct{ a, b, want int }{-10, 3, -3},
		struct{ a, b, want int }{-1, 3, 0},
		struct{ a, b, want int }{0, 3, 0},
		struct{ a, b, want int }{1, 3, 1},
		struct{ a, b, want int }{10, 3, 4},
		struct{ a, b, want int }{-10, -1, 10},
		struct{ a, b, want int }{-1, -1, 1},
		struct{ a, b, want int }{0, -1, 0},
		struct{ a, b, want int }{1, -1, -1},
		struct{ a, b, want int }{10, -1, -10},
		struct{ a, b, want int }{-10, -2, 5},
		struct{ a, b, want int }{-1, -2, 1},
		struct{ a, b, want int }{0, -2, 0},
		struct{ a, b, want int }{1, -2, 0},
		struct{ a, b, want int }{10, -2, -5},
		struct{ a, b, want int }{-10, -3, 4},
		struct{ a, b, want int }{-1, -3, 1},
		struct{ a, b, want int }{0, -3, 0},
		struct{ a, b, want int }{1, -3, 0},
		struct{ a, b, want int }{10, -3, -3},
	}

	for _, test := range tests {
		if got := divideCeil(test.a, test.b); got != test.want {
			t.Errorf("divideCeil(%v, %v) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}
