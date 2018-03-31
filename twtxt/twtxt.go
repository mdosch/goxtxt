package twtxt

import (
	"log"
	"os"
	"os/exec"
	"strconv"
)

var shell string = initshell()

func initshell() string {
	shell := os.Getenv("SHELL")
        if _, err := os.Stat(shell); os.IsNotExist(err) {
                log.Fatal("Error: ", err)
        }
	return shell
}

func Tweet(twtxtpath *string, s []string) {
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

func Timeline(twtxtpath *string, i *int) *string {
	command := *twtxtpath + " timeline | head -n " + strconv.Itoa(*i*3)
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		log.Fatal(err)
	}
	outputstring := string(out)
	return &outputstring 
}

func ViewUser(twtxtpath *string, i *int, user *string) *string {
	command := *twtxtpath + " view " + *user + " | head -n " + strconv.Itoa(*i*3)
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		log.Fatal(err)
	}
	outputstring := string(out)
	return &outputstring
}

func UserManagement(twtxtpath *string, follow bool, s []string) *string {
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

func Listfollowing(twtxtpath *string) *string {
	command := *twtxtpath + " following"
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		log.Fatal(err)
	}
	outputstring := string(out)
	return &outputstring
}

func Mentions(twtxtpath *string, nick *string, number *int) *string {
	command := *twtxtpath + " timeline | grep -B1 -m " + strconv.Itoa(*number) + " @" + *nick + " "
	out, err := exec.Command(shell, "-c", command).Output()
	if err != nil {
		outputstring := "No mentions found."
		return &outputstring
	}
	outputstring := string(out)
	return &outputstring
}
