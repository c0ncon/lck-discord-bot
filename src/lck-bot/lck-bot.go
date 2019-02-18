package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	token            string
	schedules        []schedule
	weekdayKor       = [...]string{"일", "월", "화", "수", "목", "금", "토"}
	imgRespRegexp, _ = regexp.Compile("^\\(([\\w\\d\\s가-힣]+)\\)$")
	imageURLs        = map[string]string{}
)

type schedule struct {
	Date    string  `json:"date"`
	Matches []match `json:"matches"`
}

type match []string

func init() {
	loadToken(&token)
	loadSchedules("schedules.json", &schedules)
	loadImageURLs(&imageURLs)
}

func main() {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Panicln("error creating Discord session,", err)
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		log.Panicln("error opening connection,", err)
	}

	user, err := dg.User("@me")
	if err != nil {
		log.Panicln("error opening connection,", err)
	}
	log.Println("Bot is now running. Press CTRL-C to exit.")
	log.Printf("Bot invite URL:\n\thttps://discordapp.com/oauth2/authorize?client_id=%s&scope=bot\n", user.ID)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	matched, t := stringMatch(m.Content)
	if matched {
		c, _ := s.State.Channel(m.ChannelID)
		g, _ := s.State.Guild(c.GuildID)
		log.Printf("%s: %s@%s#%s\n", t, m.Author.Username, g.Name, c.Name)

		switch t {
		case "s":
			s.ChannelMessageSend(m.ChannelID, getNextMatch())
			return
		case "i":
			str := imgRespRegexp.FindStringSubmatch(m.Content)[1]
			if imageURLs[str] != "" {
				s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
					Image: &discordgo.MessageEmbedImage{
						URL: imageURLs[str],
					},
				})
			}
			return
		}
	}
}

func stringMatch(str string) (bool, string) {
	if str == "뜬뜬" || str == "ㄸㄸ" {
		return true, "s"
	} else if imgRespRegexp.MatchString(str) {
		return true, "i"
	} else {
		return false, ""
	}
}

func loadToken(token *string) {
	var t string
	if _, err := os.Stat(".token"); err == nil {
		tokenFile, _ := os.Open(".token")
		defer tokenFile.Close()
		scanner := bufio.NewScanner(tokenFile)
		scanner.Scan()
		t = scanner.Text()
	} else {
		flag.StringVar(&t, "t", "", "Bot Token")
		flag.Parse()
		f, err := os.Create(".token")
		if err != nil {
			log.Fatalln("error creating file,", err)
		}
		defer f.Close()
		f.WriteString(t)
	}
	*token = t
}

func loadSchedules(path string, sc *[]schedule) {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Panicln("error opening file,", err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Panicln(err)
	}
	json.Unmarshal(byteValue, sc)
}

func loadImageURLs(urls *map[string]string) {
	if _, err := os.Stat("imageurls.json"); err == nil {
		jsonFile, err := os.Open("imageurls.json")
		if err != nil {
			log.Fatalln(err)
		}
		defer jsonFile.Close()

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			log.Fatalln(err)
		}
		json.Unmarshal(byteValue, urls)
	}
}

func getNextMatch() string {
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 00, 00, 00, 00, time.Local)
	nextMatch := "잘몰르겠음 몬가.. 몬가 일어나고잇음"

	for _, sc := range schedules {
		t, _ := time.Parse("2006-01-02", sc.Date)

		if t.After(today) {
			matches := make([]string, len(sc.Matches))
			for i, m := range sc.Matches {
				matches[i] = strings.Join(m, " vs ")
			}

			nextMatch = sc.Date + "(" + weekdayKor[t.Weekday()] + "): " + strings.Join(matches, " / ")
			break
		}
	}

	return nextMatch
}
