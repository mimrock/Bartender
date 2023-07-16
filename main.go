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
	log.WithField("configFile", configFile).Info("Bartender v0.2 starting up.")

	cfg, err := config.NewConfig(configFile)
	if err != nil {
		log.Fatal("Cannot load config:", err.Error())
	}

	setLogLevel(cfg.LogLevel)

	rock, err := rocket.NewConnectionFromConfig(cfg)

	if err != nil {
		log.Fatal("Cannot create new rocketchat connection:", err.Error())
	}

	log.WithField("rock", rock).Debug("Connection to rocketchat established.")

	err = rock.UserTemporaryStatus(rocket.STATUS_ONLINE)
	if err != nil {
		log.WithError(err).Error("Cannot set temporary status to online.")
	}
	//rock.UserDefaultStatus(rocket.STATUS_ONLINE)

	oa := openai.NewFromConfig(cfg)

	hist := NewHistoryFromConfig(cfg)

	for {
		// Wait for a new message to come in
		msg, err := rock.GetNewMessage()

		// If error, quit because that means the connection probably quit
		if err != nil {
			break
		}

		// If begins with '@Username ' or is in private chat
		// @todo robot must be pinged in a private room
		if msg.AmIPinged || msg.IsDirect {
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
