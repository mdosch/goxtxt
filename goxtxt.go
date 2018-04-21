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
	"fluux.io/xmpp"
	"fmt"
	"github.com/mdosch/goxtxt/twtxt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var lastActivity = time.Now()

func main() {

	type Configuration struct {
		Address         string
		BotJid          string
		Password        string
		ControlJid      string
		Twtxtnick       string
		TimelineEntries int
		MaxCharacters   int
	}

	var err error
	configpath := os.Getenv("HOME") + "/.config/goxtxt/"
	if _, err := os.Stat(configpath + "config.json"); os.IsNotExist(err) {
		err = os.MkdirAll(configpath, 0700)
		if err != nil {
			log.Fatal("Error: ", err)
		}
	}
	if _, err := os.Stat(configpath + "config.json"); os.IsNotExist(err) {
		log.Fatal("Error: ", err)
	}

	file, _ := os.Open(configpath + "config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	if err := decoder.Decode(&configuration); err != nil {
		log.Fatal("Error: ", err)
	}

	options := xmpp.Options{
		Address:  configuration.Address,
		Jid:      configuration.BotJid,
		Password: configuration.Password,
		//		PacketLogger: os.Stdout,
		Insecure: false}

	var client *xmpp.Client
	if client, err = xmpp.NewClient(options); err != nil {
		log.Fatal("Error: ", err)
	}

	var session *xmpp.Session

	// Connecting to xmpp server
	if session, err = client.Connect(); err != nil {
		log.Fatal("Error: ", err)
	}

	fmt.Println("Stream opened, we have streamID = ", session.StreamId)

	// Starting goroutine to check in background if connection is still alive.
	go checkConnection(client, &configuration.BotJid, &configuration.Address)

	var words []string

	// Receiving xmpp packets in a for loop
	for packet := range client.Recv() {
		switch packet := packet.(type) {
		case xmpp.Message:
			lastActivity = time.Now()
			if strings.HasPrefix(packet.From, configuration.ControlJid) == false {
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "You're not allowed to control me."}
				client.Send(reply)
				fmt.Fprintf(os.Stdout, "Body = %s - from = %s\n", packet.Body, packet.From)
				break
			}
			words = strings.Fields(packet.Body)
			switch strings.ToLower(words[0]) {
			case "help":
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From,
					Type: "chat"}, Body: "\"help\": Show this message.\n" +
					"\"ping\": Bot replies if available.\n" +
					"\"tl\": Show last " + strconv.Itoa(configuration.TimelineEntries) +
					" timeline entries.\n" +
					"\"tv [user]\": Show [user]s timeline.\n" +
					"\"tw [tweet]\": Will tweet your input [tweet] and afterwards show your timeline.\n" +
					"\"tm [user]\": Will show the last " + strconv.Itoa(configuration.TimelineEntries) +
					" mentions. [user] will fall back  to \"" + configuration.Twtxtnick + "\" if not specified.\n" +
					"\"tt [tag]\": Will show the last " + strconv.Itoa(configuration.TimelineEntries) +
					" occurrences of #[tag]\n" +
					"\"tf [user] [url]\": Follow [user].\n" +
					"\"tu [user]\": Unfollow [user].\n" +
					"\"to\": List the accounts you are following.\n" +
					"\"source\": Shows a link to the sourcecode."}
				client.Send(reply)
			case "ping":
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Pong!"}
				client.Send(reply)
			case "source":
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "https://github.com/mdosch/goxtxt/"}
				client.Send(reply)
			case "tw":
				if len(words) == 1 {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "No Input."}
					client.Send(reply)
					break
				}
				if len(packet.Body)-3 > configuration.MaxCharacters {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Tweet exceeds maximum of " +
						strconv.Itoa(configuration.MaxCharacters) + " characters."}
					client.Send(reply)
					break
				}
				_, err := twtxt.Tweet(words[1:])
				if err != nil {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Failed."}
					client.Send(reply)
					break
				}
				fallthrough
			case "tl":
				out, err := twtxt.Timeline(&configuration.TimelineEntries)
				if err != nil {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Failed."}
					client.Send(reply)
					break
				}
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: *out}
				client.Send(reply)
			case "tv":
				if len(words) == 1 {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "No Input."}
					client.Send(reply)
					break
				}
				if len(words) > 2 {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Timeline view supports only one user."}
					client.Send(reply)
					break
				}
				out, err := twtxt.ViewUser(&configuration.TimelineEntries, &words[1])
				if err != nil {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Failed."}
					client.Send(reply)
					break
				}
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: *out}
				client.Send(reply)
			case "tm":
				if len(words) == 1 {
					out, err := twtxt.Mentions(&configuration.Twtxtnick, &configuration.TimelineEntries)
					if err != nil {
						reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Failed."}
						client.Send(reply)
						break
					}
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: *out}
					client.Send(reply)
				}
				if len(words) == 2 {
					out, err := twtxt.Mentions(&words[1], &configuration.TimelineEntries)
					if err != nil {
						reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Failed."}
						client.Send(reply)
						break
					}
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: *out}
					client.Send(reply)
				}
				if len(words) > 2 {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Too many arguments."}
					client.Send(reply)
				}
			case "tt":
				if len(words) == 1 {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Missing Input."}
					client.Send(reply)
				}
				if len(words) == 2 {
					out, err := twtxt.Tags(&words[1], &configuration.TimelineEntries)
					if err != nil {
						reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Failed."}
						client.Send(reply)
						break
					}
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: *out}
					client.Send(reply)
				}
				if len(words) > 2 {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Too many arguments."}
					client.Send(reply)
				}
			case "tf":
				if len(words) != 3 {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Missing Input."}
					client.Send(reply)
					break
				}
				out, err := twtxt.UserManagement(true, words[1:])
				if err != nil {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Failed."}
					client.Send(reply)
					break
				}
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: *out}
				client.Send(reply)
			case "tu":
				if len(words) != 2 {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Wrong parameter count."}
					client.Send(reply)
					break
				}
				out, err := twtxt.UserManagement(false, words[1:])
				if err != nil {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Failed."}
					client.Send(reply)
					break
				}
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: *out}
				client.Send(reply)
			case "to":
				out, err := twtxt.ListFollowing()
				if err != nil {
					reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Failed."}
					client.Send(reply)
					break
				}
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: *out}
				client.Send(reply)
			default:
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"}, Body: "Unknown command. Send \"help\"."}
				client.Send(reply)
			}
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
