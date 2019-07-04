package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	token            string
	matches          []match
	matchMap         = map[string][]string{}
	dates            []string
	weekdayKor       = [...]string{"일", "월", "화", "수", "목", "금", "토"}
	imgRespRegexp, _ = regexp.Compile("^\\(([\\w\\d\\s가-힣]+)\\)$")
	imageURLs        = map[string]string{}
)

type match struct {
	Date string `json:"date"`
	Time string `json:"time"`
	Home string `json:"home"`
	Away string `json:"away"`
}

func init() {
	loadToken(&token)
	loadSchedules("schedules.json", &matches)
	makeScheduleMap(matches, matchMap, &dates)
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

		switch t {
		case "s":
			log.Printf("%s: %s@%s#%s\n", t, m.Author.Username, g.Name, c.Name)
			s.ChannelMessageSend(m.ChannelID, getNextMatch())
			return
		case "i":
			str := imgRespRegexp.FindStringSubmatch(m.Content)[1]
			if imageURLs[str] != "" {
				log.Printf("%s: %s@%s#%s\n", t, m.Author.Username, g.Name, c.Name)
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

func loadSchedules(path string, matches *[]match) {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Panicln("error opening file,", err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Panicln(err)
	}
	json.Unmarshal(byteValue, matches)
}

func makeScheduleMap(matches []match, matchMap map[string][]string, dates *[]string) {
	for _, match := range matches {
		m := fmt.Sprintf("%s\t%-15svs%15s", match.Time, match.Home, match.Away)
		matchMap[match.Date] = append(matchMap[match.Date], m)
	}
	for date := range matchMap {
		*dates = append(*dates, date)
	}
	sort.Strings(*dates)
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

	for _, date := range dates {
		t, _ := time.Parse("2006-01-02", date)
		if t.Equal(today) || t.After(today) {
			nextMatch = fmt.Sprintf("```%s\n\n%s```", date+"("+weekdayKor[t.Weekday()]+")", strings.Join(matchMap[date], "\n"))
			break
		}
	}

	return nextMatch
}
