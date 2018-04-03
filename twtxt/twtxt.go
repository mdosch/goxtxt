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

package twtxt

import (
	"log"
	"os"
	"os/exec"
	"strconv"
)

var shell string = initshell()

/* Needed as workaround as exec.Command fails when using e. g.
"/usr/local/bin/twtxt" as comamnd and "timeline | head -n 30"
as argument but works well when using "/bin/zsh" as command
and passing "-c" and "/usr/local/bin/twtxt timeline | head -n 30"
as argument. */

func initshell() string {
	shell := os.Getenv("SHELL")
	if _, err := os.Stat(shell); os.IsNotExist(err) {
		log.Fatal("Error: ", err)
	}
	return shell
}

func Tweet(twtxtpath *string, s []string) (*string, error) {
	command := *twtxtpath + " tweet \""
	for i, tweet := range s {
		command = command + tweet
		if i < len(s)-1 {
			command = command + " "
		}
	}
	command = command + "\""
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}

func Timeline(twtxtpath *string, i *int) (*string, error) {
	command := *twtxtpath + " timeline | head -n " + strconv.Itoa(*i*3)
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}

func ViewUser(twtxtpath *string, i *int, user *string) (*string, error) {
	command := *twtxtpath + " view " + *user + " | head -n " + strconv.Itoa(*i*3)
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}

func UserManagement(twtxtpath *string, follow bool, s []string) (*string, error) {
	var command string
	if follow == true {
		command = *twtxtpath + " follow -f " + s[0] + " " + s[1]
	} else {
		command = *twtxtpath + " unfollow " + s[0]
	}
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}

func Listfollowing(twtxtpath *string) (*string, error) {
	command := *twtxtpath + " following"
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}

func Mentions(twtxtpath *string, nick *string, number *int) (*string, error) {
	command := *twtxtpath + " timeline | grep -iB1 -m " + strconv.Itoa(*number) + " \"@" + *nick + "\""
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}

func Tags(twtxtpath *string, tag *string, number *int) (*string, error) {
	command := *twtxtpath + " timeline | grep -iB1 -m " + strconv.Itoa(*number) + " \"#" + *tag + "\""
	out, err := exec.Command(shell, "-c", command).Output()
	outputstring := string(out)
	return &outputstring, err
}
