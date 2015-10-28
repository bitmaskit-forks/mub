package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func parsecommand(line string) {
	fields := strings.Fields(line)

	switch fields[0] {
	case "/nick":
		conn.Nick(fields[1])
	case "/join":
		if len(fields) != 2 {
			commanderror("Use /join #channel\n")
			return
		}

		currtarget = fields[1]
		conn.Join(currtarget)
	case "/part":
		if len(fields) != 2 {
			commanderror("Use /part #channel\n")
			return
		}

		conn.Part(fields[1])
		currtarget = ""
	case "/names":
		namescmd := fmt.Sprintf("NAMES %v", currtarget)
		conn.Raw(namescmd)
	case "/whois":
		if len(fields) != 2 {
			commanderror("Use /whois <nick>\n")
			return
		}

		conn.Whois(fields[1])

	case "/quit":
		iquit()
		if len(fields) == 2 {
			conn.Quit(fields[1])
		} else {
			conn.Quit()
		}
		quitclient = true
	}
}

func ui(sub bool) {
	quitclient = false
	for !quitclient {
		if !sub {
			fmt.Printf("[%v] ", currtarget)
		}

		bio := bufio.NewReader(os.Stdin)
		line, err := bio.ReadString('\n')
		logmsg(time.Now(), conn.Me().Nick, currtarget, line)

		if err != nil {
			log.Fatal("Couldn't get input.\n")
		}

		if line != "\n" && line != "\r\n" {
			if line[0] == '/' {
				// A command
				parsecommand(line)
			} else {
				// Send line to target.
				if currtarget == "" {
					notarget()
				} else {
					conn.Privmsg(currtarget, line)
				}
			}
		}
	}
}