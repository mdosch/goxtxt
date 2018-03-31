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
	"github.com/processone/gox/xmpp"
	"goxtxt/twtxt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {

	type Configuration struct {
		Address         string
		BotJid          string
		Password        string
		Twtxtpath       string
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
	//	options := xmpp.Options{Address: configuration.Address, Jid: configuration.BotJid, Password: configuration.Password, PacketLogger: os.Stdout}
	options := xmpp.Options{
		Address:  configuration.Address,
		Jid:      configuration.BotJid,
		Password: configuration.Password}

	if _, err := os.Stat(configuration.Twtxtpath); os.IsNotExist(err) {
		log.Fatal("Error: ", err)
	}

	var client *xmpp.Client
	if client, err = xmpp.NewClient(options); err != nil {
		log.Fatal("Error: ", err)
	}

	var session *xmpp.Session

	for { // Will this loop be enough for reconnecting after connection loss?

		if session, err = client.Connect(); err != nil {
			log.Fatal("Error: ", err)
		}

		fmt.Println("Stream opened, we have streamID = ", session.StreamId)

		var words []string

		for packet := range client.Recv() {
			switch packet := packet.(type) {
			case *xmpp.ClientMessage:
				if strings.HasPrefix(packet.From, configuration.ControlJid) {
					words = strings.Fields(packet.Body)
					switch strings.ToLower(words[0]) {
					case "help":
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "\"help\": Show this message.\n" +
							"\"ping\": Bot replies if available.\n" +
							"\"tl\": Show last " + strconv.Itoa(configuration.TimelineEntries) + " timeline entries.\n" +
							"\"tv [user]\": Show [user]s timeline.\n" +
							"\"tw [tweet]\": Will tweet your input [tweet] and afterwards show your timeline.\n" +
							"\"tm [user]\": Will show the last " + strconv.Itoa(configuration.TimelineEntries) +
							" mentions. [user] will fall back  to \"" + configuration.Twtxtnick + "\" if not specified.\n" +
							"\"tf [user] [url]\": Follow [user].\n" +
							"\"tu [user]\": Unfollow [user].\n" +
							"\"to\": List the accounts you are following.\n" +
							"\"source\": Shows a link to the sourcecode."}
						client.Send(reply.XMPPFormat())
					case "ping":
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "Pong!"}
						client.Send(reply.XMPPFormat())
					case "source":
                                                reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "https://github.com/mdosch/goxtxt/"}
                                                client.Send(reply.XMPPFormat())
					case "tw":
						if len(words) == 1 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "No Input."}
							client.Send(reply.XMPPFormat())
							break
						}
						if len(packet.Body)-3 > configuration.MaxCharacters {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "Tweet exceeds maximum of " +
								strconv.Itoa(configuration.MaxCharacters) + " characters."}
							client.Send(reply.XMPPFormat())
							break
						}
						twtxt.Tweet(&configuration.Twtxtpath, words[1:])
						fallthrough
					case "tl":
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *twtxt.Timeline(&configuration.Twtxtpath,
							&configuration.TimelineEntries)}
						client.Send(reply.XMPPFormat())
					case "tv":
						if len(words) == 1 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "No Input."}
							client.Send(reply.XMPPFormat())
							break
						}
						if len(words) > 2 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "Timeline view supports only one user."}
							client.Send(reply.XMPPFormat())
							break
						}
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *twtxt.ViewUser(&configuration.Twtxtpath,
							&configuration.TimelineEntries, &words[1])}
						client.Send(reply.XMPPFormat())
					case "tm":
						if len(words) == 1 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *twtxt.Mentions(&configuration.Twtxtpath,
								&configuration.Twtxtnick, &configuration.TimelineEntries)}
							client.Send(reply.XMPPFormat())
						}
						if len(words) == 2 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *twtxt.Mentions(&configuration.Twtxtpath,
								&words[1], &configuration.TimelineEntries)}
							client.Send(reply.XMPPFormat())
						}
						if len(words) > 2 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "Too many arguments."}
							client.Send(reply.XMPPFormat())
						}
					case "tf":
						if len(words) != 3 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "Missing Input."}
							client.Send(reply.XMPPFormat())
							break
						}
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *twtxt.UserManagement(&configuration.Twtxtpath, true, words[1:])}
						client.Send(reply.XMPPFormat())
					case "tu":
						if len(words) != 2 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "Wrong parameter count."}
							client.Send(reply.XMPPFormat())
							break
						}
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *twtxt.UserManagement(&configuration.Twtxtpath, false, words[1:])}
						client.Send(reply.XMPPFormat())
					case "to":
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *twtxt.Listfollowing(&configuration.Twtxtpath)}
						client.Send(reply.XMPPFormat())
					default:
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "Unknown command. Send \"help\"."}
						client.Send(reply.XMPPFormat())
					}
				} else {
					reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "You're not allowed to control me."}
					client.Send(reply.XMPPFormat())
					fmt.Fprintf(os.Stdout, "Body = %s - from = %s\n", packet.Body, packet.From)
				}
			default:
				fmt.Fprintf(os.Stdout, "Ignoring packet: %T\n", packet)

			}
		}
		fmt.Fprintf(os.Stdout, "Reconnecting")
	}
}
