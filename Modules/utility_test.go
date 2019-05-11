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
