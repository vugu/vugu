package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vugu/vugu"
)

type Root struct {
	bpi       bpi
	isLoading bool
}

type bpi struct {
	Time struct {
		Updated string `json:"updated"`
	} `json:"time"`
	BPI map[string]struct {
		Code      string  `json:"code"`
		Symbol    string  `json:"symbol"`
		RateFloat float64 `json:"rate_float"`
	} `json:"bpi"`
}

var c Root

func (c *Root) HandleClick(event vugu.DOMEvent) {

	c.bpi = bpi{}

	go func(ee vugu.EventEnv) {

		ee.Lock()
		c.isLoading = true
		ee.UnlockRender()

		res, err := http.Get("https://api.coindesk.com/v1/bpi/currentprice.json")
		if err != nil {
			log.Printf("Error fetch()ing: %v", err)
			return
		}
		defer res.Body.Close()

		var newb bpi
		err = json.NewDecoder(res.Body).Decode(&newb)
		if err != nil {
			log.Printf("Error JSON decoding: %v", err)
			return
		}

		ee.Lock()
		defer ee.UnlockRender()
		c.bpi = newb
		c.isLoading = false

	}(event.EventEnv())
}
