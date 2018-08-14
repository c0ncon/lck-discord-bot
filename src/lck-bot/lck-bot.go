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

var (
	Token      string
	schedules  Schedules
	WeekdayKor = [...]string{"일", "월", "화", "수", "목", "금", "토"}
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
}

func main() {
	Token = ""
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "뜬뜬" || m.Content == "ㄸㄸ" {
		t, _ := time.Parse(time.RFC3339, string(m.Timestamp))
		t = t.Add(9 * time.Hour)
		c, _ := s.State.Channel(m.ChannelID)
		g, _ := s.State.Guild(c.GuildID)
		fmt.Printf("%s %s@%s#%s\n", t.Format("2006-01-02 15:04"), m.Author.Username, g.Name, c.Name)
		s.ChannelMessageSend(m.ChannelID, getNextMatch())
	}
}

func loadSchedules(path string, sc *Schedules) {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println("error opening file,", err)
		os.Exit(1)
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

			nextMatch = sc.Date + "(" + WeekdayKor[t.Weekday()] + "): " + strings.Join(matches, " / ")
			break
		}
	}

	return nextMatch
}
