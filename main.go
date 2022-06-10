package main

import (
  "log"
  "os"
  "os/signal"
	"syscall"
  "strings"
  "net/http"
  "net/url"
  "flag"
  "encoding/json"
  "github.com/bwmarrin/discordgo"
  "fmt"
)

var (
  Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
  // create a discordgo object or exit with an error
  disc, err := discordgo.New("Bot " + Token)
  	if err != nil {
		log.Fatalln("Error creating DiscordGo object, check your token,", err)
    os.Exit(1)
	}

  // register the "respond" method as the handler for new messages
  disc.AddHandler(respond)

  // tell discord handler to only look for message events
  disc.Identify.Intents = discordgo.IntentsGuildMessages

  // connect to discord and listen to messages
  err = disc.Open()
  if err != nil {
    log.Fatalln("Error creating connection to Discord")
    os.Exit(1)
  }

  // wait for exit command
  log.Println("Old Head Bot is active and listening.")
  sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

  // close the listener
  disc.Close()
}

// method to determine the respose to any message
func respond(s *discordgo.Session, m *discordgo.MessageCreate) {
  // ignore messages from the bot
  if m.Author.Bot {
		return
	}

  // ignore messages that do not start with "!"
  if ! strings.HasPrefix(m.Content, "!") { 
    return
  }

  // if string is "!define" return the top urbandictionary definition and example
  if strings.HasPrefix(m.Content, "!define ") {
    word := strings.TrimPrefix(m.Content, "!define ")
    log.Println("Getting Definition for: " + word)
    s.ChannelMessageSend(m.ChannelID, define(word))
  }
}

// structs required to decode JSON
type UrbanDictList struct {
  List []UrbanDictDefinition
}

type UrbanDictDefinition struct {
  Definition string
  Example string
}

// find and return the definition from urban dictionary
func define(word string) string {
  // attempt a get from urbandictionary API
  resp, err := http.Get("https://api.urbandictionary.com/v0/define?term=" + url.QueryEscape(word))
  if err != nil {
    log.Fatalln(err)
  }

  // place the json in a struct
  var definitions UrbanDictList
  err = json.NewDecoder(resp.Body).Decode(&definitions)
  if err != nil {
    log.Fatalln(err)
  }

  // if json is empty, return a different string
  if len(definitions.List) == 0 {
    return "Congrats man, I literally have no idea what `" + word + "` is."
  }

  // get definition and example but remove brackets
  def := strings.ReplaceAll(definitions.List[0].Definition, "[", "")
  def = strings.ReplaceAll(def, "]", "")

  ex := strings.ReplaceAll(definitions.List[0].Example, "[", "")
  ex = strings.ReplaceAll(ex, "]", "")

  // return formatted string
  return fmt.Sprintf("I gotchu fam.\n\n**Definition:**\n```text\n%s\n```\n**Example:**\n```text\n%s\n```", def, ex)
}
