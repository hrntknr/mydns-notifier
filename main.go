package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/comail/colog"
	"github.com/robfig/cron"
	"gopkg.in/xmlpath.v2"
)

type config struct {
	Notice notice
	Log    logger
}

type notice struct {
	ID       string
	Password string
	IPv4     bool
	IPv6     bool
	Cron     string
}

type logger struct {
	Slack slack
}

type slack struct {
	Enable  bool
	HookURL string
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.toml", "configPath")
	var verbosity int
	flag.IntVar(&verbosity, "verbosity", 3, "verbosity, 1,2,3,4,5")
	flag.Parse()
	exitCode := notifierMain(configPath, verbosity)
	os.Exit(exitCode)
}

func notifierMain(configPath string, verbosity int) int {
	switch verbosity {
	case 1:
		colog.SetMinLevel(colog.LError)
		break
	case 2:
		colog.SetMinLevel(colog.LWarning)
		break
	case 3:
		colog.SetMinLevel(colog.LInfo)
		break
	case 4:
		colog.SetMinLevel(colog.LDebug)
		break
	case 5:
		colog.SetMinLevel(colog.LTrace)
		break
	default:
		flag.Usage()
		return 2
	}
	colog.SetFormatter(&colog.StdFormatter{
		Colors: true,
		Flag:   log.Ldate | log.Ltime | log.Lshortfile,
	})
	colog.Register()

	var config config
	initConfig(&config)
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		fmt.Print(err.Error())
		return 1
	}
	if config.Notice.Cron == "" {
		log.Print("debug: run once mode.")
		return notify(config)
	}
	log.Print("debug: run cron mode.")
	c := cron.New()
	c.AddFunc(config.Notice.Cron, func() {
		notify(config)
	})
	c.Start()
	defer c.Stop()
	select {}
}

func initConfig(config *config) {
	config.Notice.Cron = ""
	config.Notice.ID = ""
	config.Notice.Password = ""
	config.Notice.IPv4 = true
	config.Notice.IPv6 = true
	config.Log.Slack.Enable = false
	config.Log.Slack.HookURL = ""
}

func notify(config config) int {
	var exitCode = 0
	if config.Notice.IPv4 {
		ip, err := notifyIP("ipv4", config.Notice.ID, config.Notice.Password)
		if err == nil {
			logResult(ip, config.Notice.ID, config.Log)
		} else {
			logError(err, config.Notice.ID, config.Log)
			exitCode = 1
		}
	}
	if config.Notice.IPv6 {
		ip, err := notifyIP("ipv6", config.Notice.ID, config.Notice.Password)
		if err == nil {
			logResult(ip, config.Notice.ID, config.Log)
		} else {
			logError(err, config.Notice.ID, config.Log)
			exitCode = 1
		}
	}
	return exitCode
}

func notifyIP(protocol string, id string, password string) (string, error) {
	log.Print("debug: running notify.")
	client := &http.Client{}
	url := fmt.Sprintf("https://%s.mydns.jp/login.html", protocol)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return "", err
	}

	req.SetBasicAuth(id, password)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return "", err
	}
	log.Printf("trace: POST %s success", url)

	node, err := xmlpath.ParseHTML(res.Body)
	if err != nil {
		log.Printf("error: %s", err.Error())
		return "", err
	}

	if res.StatusCode != 200 {
		path := xmlpath.MustCompile("/html/head/title")
		title, _ := path.String(node)
		err := fmt.Errorf("%s", title)
		log.Printf("error: %s", err.Error())
		return "", err
	}
	path := xmlpath.MustCompile("/html/body/dd[2]")
	ip, _ := path.String(node)
	log.Printf("info: notify %s: %s", protocol, ip)
	return ip, nil
}

type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
}
type attachment struct {
	Fallback string  `json:"fallback"`
	Color    string  `json:"color"`
	Fields   []field `json:"fields"`
}
type jsonStruct struct {
	Attachments []attachment `json:"attachments"`
}

func logResult(message string, id string, logger logger) {
	if logger.Slack.Enable {
		log.Print("debug: sending result to slack.")

		_field := field{}
		_field.Title = id
		_field.Value = message
		_attachment := attachment{}
		_attachment.Color = "#00d000"
		_attachment.Fallback = message
		_attachment.Fields = []field{_field}
		_jsonStruct := jsonStruct{}
		_jsonStruct.Attachments = []attachment{_attachment}

		jsonValue, _ := json.Marshal(_jsonStruct)
		log.Printf("trace: hook url: %s", logger.Slack.HookURL)
		log.Printf("trace: hook body: %s", bytes.NewBuffer(jsonValue))
		res, err := http.Post(logger.Slack.HookURL, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			log.Printf("error: %s", err.Error())
		}
		if res.StatusCode != 200 {
			log.Printf("error: slack status code: %d", res.StatusCode)
		}
	}
}

func logError(err error, id string, logger logger) {
	if logger.Slack.Enable {
		log.Print("debug: sending error to slack.")

		_field := field{}
		_field.Title = id
		_field.Value = err.Error()
		_attachment := attachment{}
		_attachment.Color = "#d00000"
		_attachment.Fallback = err.Error()
		_attachment.Fields = []field{_field}
		_jsonStruct := jsonStruct{}
		_jsonStruct.Attachments = []attachment{_attachment}

		jsonValue, _ := json.Marshal(_jsonStruct)
		log.Printf("trace: hook url: %s", logger.Slack.HookURL)
		log.Printf("trace: hook body: %s", bytes.NewBuffer(jsonValue))
		res, err := http.Post(logger.Slack.HookURL, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			log.Printf("error: %s", err.Error())
		}
		if res.StatusCode != 200 {
			log.Printf("error: slack status code: %d", res.StatusCode)
		}
	}
}
