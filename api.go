package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

type server struct {
	Host  string `yaml:"host"`
	Key   string `yaml:"key"`
	Poker bool   `yaml:"poker"`
}

func (s server) GetHost() string {
	if s.Host == "" {
		return "web.tada.team"
	}
	return s.Host
}

func (s server) SendMessage(text string) error {
	resp := new(struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	})
	_, err := postForm(resp, fmt.Sprintf("https://%s/api/message", s.GetHost()), url.Values{
		"key":     {s.Key},
		"message": {text},
	})
	if err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("api error: %s", resp.Error)
	}
	return nil
}

type Message struct {
	Chat              string    `json:"chat"`
	ChatDisplayName   string    `json:"chat_display_name"`
	Created           time.Time `json:"created"`
	Sender            string    `json:"sender"`
	SenderDisplayName string    `json:"sender_display_name"`
	Text              string    `json:"text"`
}

func listen(path string, rtr *mux.Router, servers []server, fn func(srv server, msg Message) error) {
	rtr.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		srv, ok := serverByKey(servers, r.Header.Get("key"))
		if !ok {
			http.Error(w, "server not found", http.StatusForbidden)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body fail", http.StatusInternalServerError)
			return
		}

		msg := Message{}
		if err := json.Unmarshal(body, &msg); len(body) != 0 && err != nil {
			http.Error(w, "unmarshal fail", http.StatusInternalServerError)
			return
		}

		if err := fn(srv, msg); err != nil {
			http.Error(w, "read body fail", http.StatusInternalServerError)
			return
		}

		io.WriteString(w, "got it")
	})
}

func serverByKey(servers []server, key string) (server, bool) {
	for _, server := range servers {
		if server.Key == key {
			return server, true
		}
	}
	return server{}, false
}
