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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func postJSON(url string, origin string, structure interface{}) (statusCode int, headers http.Header, bodyString []byte, err error) {
	jsonStr, err := json.Marshal(structure)
	if err != nil {
		return 0, nil, nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", origin)

	resp, err := myClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, resp.Header, body, nil
}

// Integer division that rounds to the next integer towards negative infinity
func divideFloor(a, b int) int {
	temp := a / b

	if ((a ^ b) < 0) && (a%b != 0) {
		return temp - 1
	}

	return temp
}

// Integer division that rounds to the next integer towards positive infinity
func divideCeil(a, b int) int {
	temp := a / b

	if ((a ^ b) >= 0) && (a%b != 0) {
		return temp + 1
	}

	return temp
}
