package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	// "os/signal"
	"regexp"
	"sort"
	"strings"

	// "syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	remoteScheduleURL = "https://raw.githubusercontent.com/c0ncon/lck-discord-bot/master/schedule.json"
	scheduleFilePath  = "./schedule.json"
	tmpSchedulePath   = "./tmp/schedule.json"
	imageURLsPath     = "./config/imageurls.json"
	configPath        = "./config/config.json"
	ucalPath          = "./config/ucal.json"
)

var (
	config           conf
	matches          []match
	matchMap         = map[string][]string{}
	dates            []string
	weekdayKor       = [...]string{"일", "월", "화", "수", "목", "금", "토"}
	imgRespRegexp, _ = regexp.Compile("^\\(([\\w\\d\\s가-힣]+)\\)$")
	imageURLs        = map[string]string{}
	ucal             []string
	ucalCount        = 0
)

type match struct {
	Date string `json:"date"`
	Time string `json:"time"`
	Home string `json:"home"`
	Away string `json:"away"`
}

type conf struct {
	Token  string   `json:"token"`
	Admins []string `json:"admins"`
}

func init() {
	config = loadConfig(configPath)
	matches = loadSchedules(scheduleFilePath)
	matchMap, dates = makeScheduleMap(matches)
	imageURLs = loadImageURLs(imageURLsPath)
	ucal = loadUcalImages(ucalPath)
}

func main() {
	dg, err := discordgo.New("Bot " + config.Token)
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

	dg.UpdateGameStatus(0, "잘몰르겠음 몬가.. 몬가 일어나고잇음")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	matched, t := stringMatch(m)
	if matched {
		c, _ := s.State.Channel(m.ChannelID)
		g, _ := s.State.Guild(c.GuildID)

		switch t {
		case "next_match":
			log.Printf("%s: %s@%s#%s\n", t, m.Author.Username, g.Name, c.Name)
			s.ChannelMessageSend(m.ChannelID, getNextMatch())
			return
		case "weekly_match":
			log.Printf("%s: %s@%s#%s\n", t, m.Author.Username, g.Name, c.Name)
			s.ChannelMessageSend(m.ChannelID, getNextWeeklyMatch())
			return
		case "image_response":
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
		case "ucal":
			log.Printf("%s: %s@%s#%s\n", t, m.Author.Username, g.Name, c.Name)
			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Image: &discordgo.MessageEmbedImage{
					URL: ucal[ucalCount],
				},
			})
			ucalCount = (ucalCount + 1) % len(ucal)
			return
		case "update":
			log.Printf("%s: %s@%s#%s\n", t, m.Author.Username, g.Name, c.Name)
			s.ChannelMessageSend(m.ChannelID, "ㅅㅅ")
			return
		}
	}
}

func stringMatch(message *discordgo.MessageCreate) (bool, string) {
	str := message.Content
	if str == "뜬뜬" || str == "ㄸㄸ" {
		return true, "next_match"
	} else if str == "!weekly" {
		return true, "weekly_match"
	} else if imgRespRegexp.MatchString(str) {
		return true, "image_response"
	} else if str == "!유칼" {
		return true, "ucal"
	} else if str == "!update" && isAdmin(message.Author.ID) {
		downloadSchedule(remoteScheduleURL)
		if isScheduleChanged(scheduleFilePath, tmpSchedulePath) {
			updateSchedule(tmpSchedulePath, scheduleFilePath)
			return true, "update"
		}
		return false, ""
	} else {
		return false, ""
	}
}

func loadConfig(path string) conf {
	var config conf
	file, err := os.Open(path)
	if err != nil {
		log.Panicln("error opening config,", err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panicln(err)
	}
	json.Unmarshal(byteValue, &config)

	return config
}

func loadSchedules(path string) []match {
	var matches []match
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Panicln("error opening file,", err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Panicln(err)
	}
	json.Unmarshal(byteValue, &matches)

	return matches
}

func makeScheduleMap(matches []match) (map[string][]string, []string) {
	var matchMap = make(map[string][]string)
	var dates []string
	for _, match := range matches {
		m := fmt.Sprintf("%-8s%-8svs%8s", match.Time, match.Home, match.Away)
		matchMap[match.Date] = append(matchMap[match.Date], m)
	}
	for date := range matchMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	return matchMap, dates
}

func loadImageURLs(path string) map[string]string {
	var urls = make(map[string]string)

	if _, err := os.Stat(path); err == nil {
		jsonFile, err := os.Open(path)
		if err != nil {
			log.Fatalln(err)
		}
		defer jsonFile.Close()

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			log.Fatalln(err)
		}
		json.Unmarshal(byteValue, &urls)
	}

	return urls
}

func loadUcalImages(path string) []string {
	var urls []string
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Panicln("error opening file", err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Panicln(err)
	}
	json.Unmarshal(byteValue, &urls)

	return urls
}

func getNextMatch() string {
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 00, 00, 00, 00, time.Local)
	nextMatch := "잘몰르겠음 몬가.. 몬가 일어나고잇음"

	for _, date := range dates {
		t, _ := time.Parse("2006-01-02", date)
		if t.Equal(today) || t.After(today) {
			nextMatch = fmt.Sprintf("```%s\n\n%s```",
				date+"("+weekdayKor[t.Weekday()]+")",
				strings.Join(matchMap[date], "\n"))
			break
		}
	}

	return nextMatch
}

func getNextWeeklyMatch() string {
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 00, 00, 00, 00, time.Local)
	var nextMatch []string

	for _, date := range dates {
		t, _ := time.Parse("2006-01-02", date)
		if t.Equal(today) || t.After(today) {
			startDay := t.AddDate(0, 0, -((int(t.Weekday()) + 6) % 7))
			for i := 0; i < 7; i++ {
				d := startDay.AddDate(0, 0, i)
				yyyymmdd := d.Format("2006-01-02")
				if match, ok := matchMap[yyyymmdd]; ok {
					nextMatch = append(
						nextMatch,
						fmt.Sprintf("```%s\n\n%s```",
							yyyymmdd+"("+weekdayKor[d.Weekday()]+")",
							strings.Join(match, "\n")))
				}
			}
			break
		}
	}

	if len(nextMatch) == 0 {
		return "잘몰르겠음 몬가.. 몬가 일어나고잇음"
	}
	return strings.Join(nextMatch, "\n")
}

func downloadSchedule(scheduleURL string) error {
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		os.Mkdir("./tmp", 755)
	}
	resp, err := http.Get(scheduleURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	newBody := strings.ReplaceAll(buf.String(), "\n", "\r\n")

	out, err := os.Create(tmpSchedulePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, strings.NewReader(newBody))
	return err
}

func isScheduleChanged(oldPath string, newPath string) bool {
	f1, err := os.Open(oldPath)
	if err != nil {
		return true
	}
	defer f1.Close()
	f2, _ := os.Open(newPath)
	defer f2.Close()

	h1 := sha256.New()
	if _, err := io.Copy(h1, f1); err != nil {
		log.Fatal(err)
	}
	h2 := sha256.New()
	if _, err := io.Copy(h2, f2); err != nil {
		log.Fatal(err)
	}
	h1Str := base64.URLEncoding.EncodeToString(h1.Sum(nil))
	h2Str := base64.URLEncoding.EncodeToString(h2.Sum(nil))

	return h1Str != h2Str
}

func updateSchedule(oldPath string, newPath string) {
	matches = nil
	for date := range matchMap {
		delete(matchMap, date)
	}
	dates = nil

	err := os.Rename(oldPath, newPath)
	if err != nil {
		log.Fatal(err)
	}
	matches = loadSchedules(scheduleFilePath)
	matchMap, dates = makeScheduleMap(matches)
}

func isAdmin(id string) bool {
	for _, admin := range config.Admins {
		if id == admin {
			return true
		}
	}

	return false
}
