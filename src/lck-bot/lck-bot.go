package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Token     string
	schedules Schedules
)

type Schedules struct {
	Schedules []Schedule `json:"schedules"`
}

type Schedule struct {
	Date    string  `json:"date"`
	Matches []Match `json:"matches"`
}

type Match []string

func init() {
	loadSchedules("schedules.json", &schedules)

	// flag.StringVar(&Token, "t", "", "Bot Token")
	// flag.Parse()
}

func main() {
	// Create a new Discord session using the provided bot token.
	Token := ""
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	if m.Content == "뜬뜬" || m.Content == "ㄸㄸ" {
		t, _ := time.Parse(time.RFC3339, string(m.Timestamp))
		t = t.Add(9 * time.Hour)
		c, _ := s.State.Channel(m.ChannelID)
		g, _ := s.State.Guild(c.GuildID)
		fmt.Printf("%s %s@%s#%s\n", t.Format("2006-01-02 15:04:06"), m.Author.Username, g.Name, c.Name)
		s.ChannelMessageSend(m.ChannelID, getNextMatch())
	}
}

func loadSchedules(path string, sc *Schedules) {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println("error opening file,", err)
		return
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, sc)
}

func getNextMatch() string {
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 00, 00, 00, 00, time.Local)
	nextMatch := "잘몰르겠음 몬가.. 몬가 일어나고잇음"

	for _, sc := range schedules.Schedules {
		t, _ := time.Parse("2006-01-02", sc.Date)

		if t.After(today) {
			matches := make([]string, len(sc.Matches))
			for i, m := range sc.Matches {
				matches[i] = strings.Join(m, " vs ")
			}

			nextMatch = sc.Date + ": " + strings.Join(matches, " / ")
			break
		}
	}

	return nextMatch
}
