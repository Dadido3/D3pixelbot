package main

import (
	rice "github.com/GeertJohan/go.rice"
)

func init() {
	rice.MustFindBox("ui")
}
