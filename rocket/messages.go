package rocket

import (
	"fmt"
	"strings"
	"time"
)

type Message struct {
	IsNew       bool                `yaml:"IsNew"`
	AmIPinged   bool                `yaml:"AmIPinged"`
	IsDirect    bool                `yaml:"IsDirect"`
	IsMention   bool                `yaml:"IsMention"`
	IsEdited    bool                `yaml:"IsEdited"`
	IsMe        bool                `yaml:"IsMe"`
	Id          string              `yaml:"Id"`
	UserName    string              `yaml:"UserName"`
	UserId      string              `yaml:"UserId"`
	RoomName    string              `yaml:"RoomName"`
	RoomId      string              `yaml:"RoomId"`
	Text        string              `yaml:"Text"`
	Timestamp   time.Time           `yaml:"Timestamp"`
	UpdatedAt   time.Time           `yaml:"UpdatedAt"`
	Reactions   map[string][]string `yaml:"Reactions"`
	Attachments []attachment        `yaml:"Attachments"`
	QuotedMsgs  []string            `yaml:"QuotedMsgs"`
	obj         map[string]interface{}
	rocketCon   *RocketCon
}

type attachment struct {
	Description string
	Title       string
	Type        string
	Link        string
}

var lastMessageTime time.Time

func init() {
	lastMessageTime = time.Now()
}

func (rock *RocketCon) handleMessageObject(obj map[string]interface{}) Message {
	var msg Message
	msg.rocketCon = rock
	msg.IsNew = true
	_, msg.IsEdited = obj["editedAt"]
	if msg.IsEdited {
		msg.IsNew = false
	}
	msg.Id = obj["_id"].(string)
	msg.Text = obj["msg"].(string)
	msg.RoomId = obj["rid"].(string)
	msg.UserId = obj["u"].(map[string]interface{})["_id"].(string)
	msg.UserName = obj["u"].(map[string]interface{})["username"].(string)

	if attachments, ok := obj["attachments"]; ok && attachments != nil {
		msg.Attachments = make([]attachment, 0)
		for _, val := range attachments.([]interface{}) {
			if _, ok := val.(map[string]interface{})["description"]; ok {
				var attach attachment
				attach.Description = val.(map[string]interface{})["description"].(string)
				attach.Title = val.(map[string]interface{})["title"].(string)
				attach.Link = val.(map[string]interface{})["title_link"].(string)
				attach.Type = val.(map[string]interface{})["type"].(string)
				msg.Attachments = append(msg.Attachments, attach)
			}
		}
	}

	if msg.UserId == rock.UserId {
		msg.IsMe = true
	}

	if strings.Contains(strings.ToLower(msg.Text), fmt.Sprintf("@%s", strings.ToLower(rock.UserName))) {
		msg.IsMention = true
	}

	if len(msg.Text) > len(rock.UserName)+2 {
		if string(strings.ToLower(msg.Text)[:len(rock.UserName)+2]) == fmt.Sprintf("@%s ", strings.ToLower(rock.UserName)) {
			msg.AmIPinged = true
		}
	}

	if _, ok := obj["unread"]; !ok {
		// Not sure what it was, but currently this key is not part of the map, and there's no repeated messages with
		// this commented out.
		//msg.IsNew = false
	}

	if _, ok := obj["urls"]; ok && len(obj["urls"].([]interface{})) != 0 {
		if _, ok := obj["urls"].([]interface{})[0].(map[string]interface{})["meta"]; ok {
			msg.IsNew = false
		}
	}

	if react, ok := obj["reactions"]; ok {
		msg.IsNew = false
		msg.Reactions = make(map[string][]string)
		for emote, val := range react.(map[string]interface{}) {
			for _, username := range val.(map[string]interface{})["usernames"].([]interface{}) {
				msg.Reactions[emote] = append(msg.Reactions[emote], username.(string))
			}
		}
	}
	if _, ok := obj["ts"]; ok {
		var unixTs float64
		switch obj["ts"].(type) {
		case map[string]interface{}:
			unixTs = obj["ts"].(map[string]interface{})["$date"].(float64)
			msg.Timestamp = time.Unix(int64(unixTs/1000), 0)
		case string:
			msg.Timestamp, _ = time.Parse("2006-01-02T15:04:05.999999999Z", obj["ts"].(string))
		}
	}

	if _, ok := obj["_updatedAt"]; ok {
		var unixTs float64
		switch obj["_updatedAt"].(type) {
		case map[string]interface{}:
			unixTs = obj["_updatedAt"].(map[string]interface{})["$date"].(float64)
			msg.Timestamp = time.Unix(int64(unixTs/1000), 0)
		case string:
			msg.Timestamp, _ = time.Parse("2006-01-02T15:04:05.999999999Z", obj["_updatedAt"].(string))
		}
	}

	if val, ok := rock.channels[msg.RoomId]; ok {
		msg.RoomName = val
		if msg.RoomName == msg.UserName {
			msg.IsDirect = true
		}
	}

	msg.QuotedMsgs = make([]string, 0)
	if rocketLinks := strings.Split(msg.Text, "]("+rock.getHttpURL()); len(rocketLinks) > 0 {
		for _, val := range rocketLinks {
			msgIdEnd := strings.IndexAny(val, "&)")
			msgIdBegin := strings.Index(val, "?msg=")
			if msgIdBegin != -1 && msgIdEnd != -1 {
				msg.QuotedMsgs = append(msg.QuotedMsgs, val[msgIdBegin+5:msgIdEnd])
			}
		}
	}

	if msg.Timestamp.After(lastMessageTime) {
		lastMessageTime = msg.Timestamp
	} else {
		msg.IsNew = false
	}

	return msg
}

func (msg *Message) Reply(text string) (Message, error) {
	return msg.rocketCon.SendMessage(msg.RoomId, text)
}

func (msg *Message) DM(text string) (Message, error) {
	if msg.IsDirect {
		return msg.Reply(text)
	}
	return msg.rocketCon.DM(msg.UserName, text)
}

func (msg *Message) KickUser() {
	msg.Reply("/kick @" + msg.UserName)
}

func (msg *Message) React(emoji string) error {
	return msg.rocketCon.React(msg.Id, emoji)
}

func (msg *Message) GetNotAddressedText() string {
	r := msg.Text
	if len(msg.Text) > 2 && msg.AmIPinged {
		r = string(strings.ToLower(msg.Text)[len(msg.rocketCon.UserName)+2:])
	}
	return r
}

func (msg *Message) EditText(text string) error {
	obj := map[string]interface{}{
		"method": "updateMessage",
		"params": []map[string]interface{}{
			map[string]interface{}{
				"_id": msg.Id,
				"rid": msg.RoomId,
				"msg": text,
			},
		},
	}

	_, err := msg.rocketCon.runMethod(obj)
	return err
}

func (msg *Message) Delete(text string) error {
	obj := map[string]interface{}{
		"method": "deleteMessage",
		"params": []map[string]interface{}{
			map[string]interface{}{
				"_id": msg.Id,
			},
		},
	}

	_, err := msg.rocketCon.runMethod(obj)
	return err
}

func (msg *Message) SetIsTyping(typing bool) error {
	obj := map[string]interface{}{
		"method": "stream-notify-room",
		"params": []interface{}{
			msg.RoomId + "/typing",
			msg.rocketCon.DisplayName,
			typing,
		},
	}
	_, err := msg.rocketCon.runMethod(obj)
	return err
}

func (msg *Message) GetQuote() string {
	var channelType string
	var proto string
	if msg.IsDirect {
		channelType = "direct"
	} else {
		channelType = "channel"
	}
	if msg.rocketCon.HostSSL {
		proto = "https"
	} else {
		proto = "http"
	}
	return fmt.Sprintf("[](%s://%s/%s/%s?msg=%s)", proto, msg.rocketCon.HostName, channelType, msg.RoomName, msg.Id)
}
