//usr/bin/go run $0 $@ ; exit
// That's a special She-bang for go

// This is a demo rocketbot in golang
// Its purpose is to showcase some features

// Specify we are the main package (the one that contains the main function)
package main

import (
	"fmt"
	"github.com/mimrock/rocketchat_openai_bot/openai"
	"os"

	"github.com/mimrock/rocketchat_openai_bot/config"
	"github.com/mimrock/rocketchat_openai_bot/rocket"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.InfoLevel)

	configFile := os.Getenv("BARTENDER_CONFIG")
	if len(configFile) == 0 {
		configFile = "config.yaml"
	}
	log.WithField("configFile", configFile).Info("Starting up.")

	config, err := config.NewConfig(configFile)
	if err != nil {
		log.Fatal("Cannot load config:", err.Error())
	}

	setLogLevel(config.LogLevel)

	rock, err := rocket.NewConnectionFromConfig(config)

	if err != nil {
		log.Fatal("Cannot create new rocketchat connection:", err.Error())
	}

	rock.UserTemporaryStatus(rocket.STATUS_ONLINE)

	oa := openai.NewFromConfig(config)

	hist := NewHistory()
	hist.Size = config.OpenAI.HistorySize
	hist.MaxLength = config.OpenAI.HistoryMaxLength

	for {
		// Wait for a new message to come in
		msg, err := rock.GetNewMessage()

		// If error, quit because that means the connection probably quit
		if err != nil {
			break
		}

		// If begins with '@Username ' or is in private chat
		// @todo robot must be pinged in a private room
		if msg.IsAddressedToMe || msg.RoomName == "" {
			log.WithField("message", msg).Debug("Incoming message for the bot.")
			err = OpenAIResponse(msg, oa, hist)
			if err != nil {
				log.WithError(err).Error("OpenAI request failed.")
				_, err = msg.Reply(fmt.Sprintf("@%s :x: Sorry, something went wrong while processing your request. This could be due to a configuration issue, a problem with the OpenAI API, or a bug in the system. Please check your configuration settings or try again later. More details can be found in the logs. :x:", msg.UserName))
				if err != nil {
					log.WithError(err).Error("Cannot send reply about the error rocketchat.")
				}

			}
		}
	}
}

func setLogLevel(logLevel string) {
	switch logLevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		log.Info("No loglevel set or invalid level. Setting loglevel to info.")
		log.SetLevel(log.InfoLevel)
	}

}
