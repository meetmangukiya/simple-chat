package common

import "encoding/json"

type Message struct {
	Op     string `json:"op"`
	Params []byte `json:"params"`
}

type SetUsernameParam = string

type MessageParam struct {
	Text     string `json:"text"`
	Username string `json:"username"`
}

func (m Message) ParseUsernameParam() (SetUsernameParam, error) {
	var username string
	err := json.Unmarshal(m.Params, &username)
	return username, err
}

func (m Message) ParseMessageParam() (MessageParam, error) {
	var message MessageParam
	err := json.Unmarshal(m.Params, &message)
	return message, err
}
