// Small cross-platform IRC client.

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	Nick     string
	RealName string
	Server   string
	LogFile  string
	TLS      bool
	Sub      bool
}

var (
	file       *os.File
	conn       *irc.Conn
	currtarget string
	quitclient bool
)

// Parse the configuration file. Returns the configuration.
func parseconfig(filename string) (conf *Config, err error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	conf = new(Config)

	if err = yaml.Unmarshal(contents, &conf); err != nil {
		return
	}

	return
}

// Log text
func logmsg(time time.Time, nick string, target string, text string) {
	line := time.UTC().Format("2006-01-02 15:04:05")
	if target != "" {
		line += " " + target
	}
	text = strings.TrimRight(text, "\r\n")
	line += " <" + nick + "> " + text + "\n"

	if file != nil {
		_, err := file.WriteString(line)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	var err error
	var configfile = flag.String("config", "mub.yaml", "Path to configuration file")

	var tlsconfig tls.Config

	flag.Parse()

	conf, err := parseconfig(*configfile)
	if err != nil {
		log.Fatal("Couldn't parse configuration file")
	}

	file, err = os.OpenFile(conf.LogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		cantopenfile(conf.LogFile, err)
	}

	cfg := irc.NewConfig(conf.Nick)
	cfg.SSL = conf.TLS
	tlsconfig.InsecureSkipVerify = true
	tlsconfig.MinVersion = tls.VersionTLS10
	cfg.SSLConfig = &tlsconfig
	cfg.Server = conf.Server
	cfg.NewNick = func(n string) string { return n + "^" }
	cfg.Me.Ident = "mub"
	cfg.Me.Name = conf.RealName
	conn = irc.Client(cfg)

	// Join channel on connect.
	conn.HandleFunc("connected",
		func(conn *irc.Conn, line *irc.Line) {
			connected(conn.Me().Nick)
		})

	conn.HandleFunc("disconnected",
		func(conn *irc.Conn, line *irc.Line) { quitclient = true })

	// Tell client to connect.
	connecting(conf.Server)

	if err := conn.Connect(); err != nil {
		connectionerror(err)
		os.Exit(-1)
	}

	conn.HandleFunc("privmsg",
		func(conn *irc.Conn, line *irc.Line) {
			msg(line.Time, line.Nick, line.Target(), line.Text())
			logmsg(line.Time, line.Nick, line.Target(), line.Text())
		})

	conn.HandleFunc("join",
		func(conn *irc.Conn, line *irc.Line) {
			joined(line.Nick, line.Target())
		})

	conn.HandleFunc("part",
		func(conn *irc.Conn, line *irc.Line) {
			parted(line.Nick, line.Target())
		})

	conn.HandleFunc("quit",
		func(conn *irc.Conn, line *irc.Line) {
			quit(line.Nick)
		})

	conn.HandleFunc("353",
		func(conn *irc.Conn, line *irc.Line) {
			members(line.Time, line.Args[2], line.Args[3])
		})

	conn.HandleFunc("311",
		func(conn *irc.Conn, line *irc.Line) {
			whois(line.Args[1], line.Args[5], line.Args[2], line.Args[3])
		})

	ui(conf.Sub)

	disconnected()
}
