package main

import (
	"fmt"
	"log"
	"time"
	"github.com/carterjones/signalr"
	"github.com/carterjones/signalr/hubs"
	"github.com/satori/go.uuid"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"strconv"
	"bytes"
	"io/ioutil"
	"encoding/base64"
	"compress/flate"
	"encoding/json"
	//"os"
)

// For more extensive use cases and capabilities, please see
// https://github.com/carterjones/bittrex.

const (
	APIKEY = ""
	APISECRET = ""
)

var c *signalr.Client

type Ticker struct {
	Symbol string
	Value  float64
}

func main() {
	createclient()
	connect()
    authenticate()
	suscribe()

	// Wait indefinitely.
	select {}
}

func createclient() {
	// Prepare a SignalR client.
	c = signalr.New(
		"socket-v3.bittrex.com",
		"1.5",
		"/signalr",
		`[{"name":"c3"}]`,
		nil,
	)
}

func connect() {
	// Start the connection.
	err := c.Run(msghandler, errorlog)
	if err != nil {
		log.Panic(err)
	}
}

func authenticate() {
	// Authenticate
	now := time.Now()   
	sec := now.Unix() * 1000
	u1 := uuid.NewV4()
	timestamp := strconv.FormatInt(sec, 10)
	uuid := fmt.Sprintf("%s", u1)
	content := timestamp + uuid
	mac := hmac.New(sha512.New, []byte(APISECRET))
	mac.Write([]byte(content))
	msg := hex.EncodeToString(mac.Sum(nil))

	arr := make([]interface{}, 0)
	arr = append(arr, APIKEY)
	arr = append(arr, sec)
	arr = append(arr, uuid)
	arr = append(arr, msg)
	err := c.Send(hubs.ClientMsg{
		H: "c3",
		M: "authenticate",
		A: arr,
		I: 1,
	})
	if err != nil {
		log.Panic(err)
	}
}

func suscribe() {
	err := c.Send(hubs.ClientMsg{
		H: "c3",
		M: "Subscribe",
		A: []interface{}{[...]string{"ticker_XRP-USD","ticker_BTC-USD","ticker_ETH-USD","ticker_DOGE-USD"}},
		I: 1,
	})
	if err != nil {
		log.Panic(err)
	}
}

func msghandler(msg signalr.Message) {
	log.Println(msg)
	for _, message := range msg.M {
		if message.M == "authenticationExpiring" {
			authenticate()
		} else {
			if message.M == "ticker" {
				z, _ := base64.StdEncoding.DecodeString(message.A[0].(string))
				r := flate.NewReader(bytes.NewReader(z))
				result, _ := ioutil.ReadAll(r)
				var tick map[string]interface{}
				json.Unmarshal(result, &tick)
				ticker := Ticker {}
				ticker.Symbol = fmt.Sprintf("%s", tick["symbol"].(string))
				ticker.Value, _ = strconv.ParseFloat(fmt.Sprintf("%s", tick["lastTradeRate"].(string)), 64)
				log.Print(ticker.Symbol)
				log.Print("->")
				log.Println(ticker.Value)
				// Save to CSV
				//t := time.Now()
				//newline := fmt.Sprintf("%s,%s,%f",t.Format("2006-01-02 15:04:05"), ticker.Symbol,ticker.Value)
				//f, _ := os.OpenFile("data/data.csv", os.O_APPEND|os.O_WRONLY, 0644)
				//fmt.Fprintln(f, newline)
				//f.Close()
			}
		}
	}
}

func errorlog (err error) {
	log.Panic(err)
}