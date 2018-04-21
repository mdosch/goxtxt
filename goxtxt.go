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

// processMessage is executing the twtxt commands according to messages
// received and replies the output.
func processMessage(client *xmpp.Client, packet *xmpp.Message, config *configuration) {
	words := strings.Fields(packet.Body)
	// First word of message body contains the command.
	switch strings.ToLower(words[0]) {
	// Show help message.
	case "help":
		reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From,
			Type: "chat"}, Body: "\"help\": Show this message.\n" +
			"\"ping\": Bot replies if available.\n" +
			"\"tl\": Show last " + strconv.Itoa(config.TimelineEntries) +
			" timeline entries.\n" +
			"\"tv [user]\": Show [user]s timeline.\n" +
			"\"tw [tweet]\": Will tweet your input [tweet] and afterwards show your timeline.\n" +
			"\"tm [user]\": Will show the last " + strconv.Itoa(config.TimelineEntries) +
			" mentions. [user] will fall back  to \"" + config.Twtxtnick + "\" if not specified.\n" +
			"\"tt [tag]\": Will show the last " + strconv.Itoa(config.TimelineEntries) +
			" occurrences of #[tag]\n" +
			"\"tf [user] [url]\": Follow [user].\n" +
			"\"tu [user]\": Unfollow [user].\n" +
			"\"to\": List the accounts you are following.\n" +
			"\"source\": Shows a link to the sourcecode."}
		client.Send(reply)
	// Reply to a ping request.
	case "ping":
		reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
			Body: "Pong!"}
		client.Send(reply)
	// Show link to source code repository.
	case "source":
		reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
			Body: "https://github.com/mdosch/goxtxt/"}
		client.Send(reply)
	// Send a tweet.
	case "tw":
		// Check that message body contains something to tweet.
		if len(words) == 1 {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "No Input."}
			client.Send(reply)
			break
		}
		// Check that tweet doesn't exceed configured maximum length.
		tweetLength := len(packet.Body) - 3
		if tweetLength > config.MaxCharacters {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Tweet exceeds maximum of " + strconv.Itoa(config.MaxCharacters) +
					" characters by " + strconv.Itoa(tweetLength-config.MaxCharacters) +
					" characters."}
			client.Send(reply)
			break
		}
		// Send the tweet.
		_, err := twtxt.Tweet(words[1:])
		if err != nil {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Failed."}
			client.Send(reply)
			break
		}
		// Automatically show updated timeline after successful tweeting.
		fallthrough
	// Show timeline.
	case "tl":
		out, err := twtxt.Timeline(&config.TimelineEntries)
		if err != nil {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Failed."}
			client.Send(reply)
			break
		}
		reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
			Body: *out}
		client.Send(reply)
	// Show only timeline entries from a certain user.
	case "tv":
		// Check there is a username specified.
		if len(words) == 1 {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "No Input."}
			client.Send(reply)
			break
		}
		// Check there is not more than one username specified.
		if len(words) > 2 {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Timeline view supports only one user."}
			client.Send(reply)
			break
		}
		out, err := twtxt.ViewUser(&config.TimelineEntries, &words[1])
		if err != nil {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Failed."}
			client.Send(reply)
			break
		}
		reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
			Body: *out}
		client.Send(reply)
	// Show @-mentions of a certain user.
	case "tm":
		// If no username specified show mentions of own user.
		if len(words) == 1 {
			out, err := twtxt.Mentions(&config.Twtxtnick, &config.TimelineEntries)
			if err != nil {
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
					Body: "Failed."}
				client.Send(reply)
				break
			}
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: *out}
			client.Send(reply)
		}
		// Show mentions of the specified user.
		if len(words) == 2 {
			out, err := twtxt.Mentions(&words[1], &config.TimelineEntries)
			if err != nil {
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
					Body: "Failed."}
				client.Send(reply)
				break
			}
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: *out}
			client.Send(reply)
		}
		// Check that there is not more than one user specified.
		if len(words) > 2 {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Too many arguments."}
			client.Send(reply)
		}
	// Show timeline entries containing a certain #-tag.
	case "tt":
		// Check that a tag is specified.
		if len(words) == 1 {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Missing Input."}
			client.Send(reply)
		}
		// Show timeline entries for the specified tag.
		if len(words) == 2 {
			out, err := twtxt.Tags(&words[1], &config.TimelineEntries)
			if err != nil {
				reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
					Body: "Failed."}
				client.Send(reply)
				break
			}
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: *out}
			client.Send(reply)
		}
		// Check that there is only one tag specified.
		if len(words) > 2 {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Too many arguments."}
			client.Send(reply)
		}
	// Add a certain user to follow.
	case "tf":
		// Check that username and URL are specified.
		if len(words) != 3 {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Missing Input."}
			client.Send(reply)
			break
		}
		// Follow the specified user.
		out, err := twtxt.UserManagement(true, words[1:])
		if err != nil {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Failed."}
			client.Send(reply)
			break
		}
		reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
			Body: *out}
		client.Send(reply)
	// Stop following a certain user.
	case "tu":
		// Check that only one username is specified.
		if len(words) != 2 {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Wrong parameter count."}
			client.Send(reply)
			break
		}
		// Unfollow the specified user.
		out, err := twtxt.UserManagement(false, words[1:])
		if err != nil {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Failed."}
			client.Send(reply)
			break
		}
		reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
			Body: *out}
		client.Send(reply)
	// Retrieve a list of users we are following.
	case "to":
		out, err := twtxt.ListFollowing()
		if err != nil {
			reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
				Body: "Failed."}
			client.Send(reply)
			break
		}
		reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
			Body: *out}
		client.Send(reply)
	// Point help command if an unknown command is received.
	default:
		reply := xmpp.Message{PacketAttrs: xmpp.PacketAttrs{To: packet.From, Type: "chat"},
			Body: "Unknown command. Send \"help\"."}
		client.Send(reply)
	}
}
