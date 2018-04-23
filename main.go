/* Copyright 2018 Martin Dosch

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License. */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"fluux.io/xmpp"
)

// configuration is defined as global as it is needed by function
// messageProcessing.
type configuration struct {
	Address         string
	BotJid          string
	Password        string
	ControlJid      string
	Twtxtnick       string
	TimelineEntries int
	MaxCharacters   int
}

// lastActivity is defined as global as it is needed by functions
// checkConnection and messageProcessing.
var lastActivity = time.Now()

func main() {

	var err error

	// Create configpath if not yet existing.
	configpath := os.Getenv("HOME") + "/.config/goxtxt/"
	if _, err := os.Stat(configpath + "config.json"); os.IsNotExist(err) {
		err = os.MkdirAll(configpath, 0700)
		if err != nil {
			log.Fatal("Error: ", err)
		}
	}

	// Check that config file is existing.
	if _, err := os.Stat(configpath + "config.json"); os.IsNotExist(err) {
		log.Fatal("Error: ", err)
	}

	// Read configuration file into variable config.
	file, _ := os.Open(configpath + "config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := configuration{}
	if err := decoder.Decode(&config); err != nil {
		log.Fatal("Error: ", err)
	}

	// Set xmpp connection options according the config.
	options := xmpp.Options{
		Address:  config.Address,
		Jid:      config.BotJid,
		Password: config.Password,
		//		PacketLogger: os.Stdout,
		Insecure: false}

	var client *xmpp.Client
	if client, err = xmpp.NewClient(options); err != nil {
		log.Fatal("Error: ", err)
	}

	var session *xmpp.Session

	// Connect to xmpp server.
	if session, err = client.Connect(); err != nil {
		log.Fatal("Error: ", err)
	}

	fmt.Println("Stream opened, we have streamID = ", session.StreamId)

	// Start goroutine to check in background if connection is still alive.
	go checkConnection(client, &config.BotJid, &config.Address)

	// Receive xmpp packets in a for loop.
	for packet := range client.Recv() {
		switch packet := packet.(type) {
		case xmpp.Message:
			lastActivity = time.Now()
			// Check if message comes from JID who is allowed to use this bot
			if strings.HasPrefix(packet.From, config.ControlJid) == false {
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
					Body: "You're not allowed to control me."}
				client.Send(reply)
				fmt.Fprintf(os.Stdout, "Body = %s - from = %s\n", packet.Body, packet.From)
				break
			}
			// Process the message.
			processMessage(client, &packet, &config)
		case xmpp.StreamError:
			fmt.Fprintf(os.Stdout, "Ignoring packet: %T\n", packet)
		default:
			lastActivity = time.Now()
		}
	}
}

// checkConnection checks every minute if there has been activity within
// the last 5 minutes and sends a ping if not.
// If there was still no activity after 2 more minutes the program
// will end with an error.
func checkConnection(client *xmpp.Client, jid *string, server *string) {
	for {
		time.Sleep(1 * time.Minute)
		timePassed := time.Since(lastActivity)
		if int(timePassed.Minutes()) >= 5.0 {
			ping := xmpp.NewIQ("get", *jid, *server, "twtxtbot", "en")
			client.Send(ping)
		}
		if int(timePassed.Minutes()) >= 7.0 {
			log.Fatal("Connection lost.")
		}
	}
}
