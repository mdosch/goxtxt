package main

import (
	"encoding/json"
	"fmt"
	"github.com/processone/gox/xmpp"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var shell string

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

	shell = os.Getenv("SHELL")
	if _, err := os.Stat(shell); os.IsNotExist(err) {
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
							"\"to\": List the accounts you are following."}
						client.Send(reply.XMPPFormat())
					case "ping":
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "Pong!"}
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
						tweet(&configuration.Twtxtpath, words[1:])
						fallthrough
					case "tl":
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *timeline(&configuration.Twtxtpath,
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
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *viewUser(&configuration.Twtxtpath,
							&configuration.TimelineEntries, &words[1])}
						client.Send(reply.XMPPFormat())
					case "tm":
						if len(words) == 1 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *mentions(&configuration.Twtxtpath,
								&configuration.Twtxtnick, &configuration.TimelineEntries)}
							client.Send(reply.XMPPFormat())
						}
						if len(words) == 2 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *mentions(&configuration.Twtxtpath,
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
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *userManagement(&configuration.Twtxtpath, true, words[1:])}
						client.Send(reply.XMPPFormat())
					case "tu":
						if len(words) != 2 {
							reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: "Wrong parameter count."}
							client.Send(reply.XMPPFormat())
							break
						}
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *userManagement(&configuration.Twtxtpath, false, words[1:])}
						client.Send(reply.XMPPFormat())
					case "to":
						reply := xmpp.ClientMessage{Packet: xmpp.Packet{To: packet.From}, Body: *listfollowing(&configuration.Twtxtpath)}
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

func tweet(twtxtpath *string, s []string) {
	command := *twtxtpath + " tweet \""
	for i, tweet := range s {
		command = command + tweet
		if i < len(s)-1 {
			command = command + " "
		}
	}
	command = command + "\""
	_, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		log.Fatal(err)
	}
}

func timeline(twtxtpath *string, i *int) *string {
	command := *twtxtpath + " timeline | head -n " + strconv.Itoa(*i*3)
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		log.Fatal(err)
	}
	outputstring := string(out)
	return &outputstring
}

func viewUser(twtxtpath *string, i *int, user *string) *string {
	command := *twtxtpath + " view " + *user + " | head -n " + strconv.Itoa(*i*3)
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		log.Fatal(err)
	}
	outputstring := string(out)
	return &outputstring
}

func userManagement(twtxtpath *string, follow bool, s []string) *string {
	var command string
	if follow == true {
		command = *twtxtpath + " follow -f " + s[0] + " " + s[1]
	} else {
		command = *twtxtpath + " unfollow " + s[0]
	}
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		log.Fatal(err)
	}
	outputstring := string(out)
	return &outputstring
}

func listfollowing(twtxtpath *string) *string {
	command := *twtxtpath + " following"
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		log.Fatal(err)
	}
	outputstring := string(out)
	return &outputstring
}

func mentions(twtxtpath *string, nick *string, number *int) *string {
	command := *twtxtpath + " timeline | grep -B1 -m " + strconv.Itoa(*number) + " @" + *nick + " "
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		outputstring := "No mentions found."
		return &outputstring
	}
	outputstring := string(out)
	return &outputstring
}
