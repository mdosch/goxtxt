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

// Package twtxt provides an interface to a local installation
// of the official twtxt client or txtnish.
package twtxt

import (
	"log"
	"os"
	"os/exec"
	"strconv"
)

var shell = initShell()
var twtxtpath, txtnish = initTwtxt()

// Needed as workaround as exec.Command fails when using e. g.
// "/usr/local/bin/twtxt" as comamnd and "timeline | head -n 30"
// as argument but works well when using "/bin/zsh" as command
// and passing "-c" and "/usr/local/bin/twtxt timeline | head -n 30"
// as argument.
func initShell() string {
	shell := os.Getenv("SHELL")
	if _, err := os.Stat(shell); os.IsNotExist(err) {
		log.Fatal("Error: ", err)
	}
	return shell
}

// Gets the path to twtxt or txtnish binary. Txtnish is prefered.
func initTwtxt() (string, bool) {
	txtnishPresent := true
	output := "/usr/local/bin/txtnish"

	if _, err := os.Stat("/usr/local/bin/txtnish"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/local/bin/twtxt"); os.IsNotExist(err) {
			log.Fatal("Error: ", err)
		}
		txtnishPresent = false
		output = "/usr/local/bin/twtxt"
	}
	return output, txtnishPresent
}

// Tweet sends a tweet.
// It returns a pointer to twtxt output and any error encountered.
func Tweet(s []string) (*string, error) {
	var command string
	for i, tweet := range s {
		command = command + tweet
		if i < len(s)-1 {
			command = command + " "
		}
	}
	out, err := exec.Command(twtxtpath, "tweet", command).Output()
	outputstring := string(out)
	return &outputstring, err
}

// Timeline shows the requested amount of timeline entries.
// It returns a pointer to the requested timeline entries and any
// error encountered.
func Timeline(entries *int) (*string, error) {
	out, err := exec.Command(twtxtpath, "timeline").Output()
	var outputstring string
	var lines int

	fullTimeline := string(out)
	for _, character := range []rune(fullTimeline) {
		outputstring += string(character)
		if string(character) == "\n" {
			lines++
		}
		if lines == *entries*3 {
			break
		}
	}
	return &outputstring, err
}

// ViewUser shows the requested amount of timeline entries for
// the specified user.
// It returns a pointer to the requested timeline entries and any
// error encountered.
func ViewUser(i *int, user *string) (*string, error) {
	if txtnish == true {
		command := twtxtpath + " timeline | grep -EiA1 " +
			"-m " + strconv.Itoa(*i) + " '^\\* " + *user + " '"
		out, err := exec.Command(shell, "-c", command).Output()
		outputstring := string(out)
		return &outputstring, err
	}
	command := twtxtpath + " view " + *user + " | head -n " + strconv.Itoa(*i*3)
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}

// UserManagement follows or unfollows the specified user.
// It returns a pointer to twtxt output and any error encountered.
func UserManagement(follow bool, s []string) (*string, error) {
	var out []byte
	var err error

	if follow == true {
		out, err = exec.Command(twtxtpath, "follow", s[0], s[1]).Output()
	} else {
		out, err = exec.Command(twtxtpath, "unfollow", s[0]).Output()
	}

	outputstring := string(out)
	println(outputstring)
	return &outputstring, err
}

// ListFollowing lists the users you are following.
// It returns a pointer to twtxt output and any error encountered.
func ListFollowing() (*string, error) {
	out, err := exec.Command(twtxtpath, "following").Output()
	outputstring := string(out)
	return &outputstring, err
}

// Mentions shows the requested amount of @-mentions for the specified user.
// It returns a pointer to the requested timeline entries and any
// error encountered.
func Mentions(nick *string, number *int) (*string, error) {
	command := twtxtpath + " timeline | grep -iB1 -m " + strconv.Itoa(*number) + " \"@" + *nick + "\""
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}

// Tags shows the requested amount of #-tags.
// It returns a pointer to the requested timeline entries and any
// error encountered.
func Tags(tag *string, number *int) (*string, error) {
	command := twtxtpath + " timeline | grep -iB1 -m " + strconv.Itoa(*number) + " \"#" + *tag + "\""
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}
