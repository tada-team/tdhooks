package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/schema"

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
		log.Println(r.Method, r.URL.RawPath)
		if r.Method != http.MethodPost {
			log.Println("method not allowed")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := r.Header.Get("key")
		srv, ok := serverByKey(servers, key)
		if !ok {
			log.Println("server not found:", key)
			http.Error(w, "server not found", http.StatusForbidden)
			return
		}

		if err := r.ParseForm(); err != nil {
			log.Println("parse form fail:", err)
		}

		if len(r.PostForm) == 0 {
			err := r.ParseMultipartForm(32 << 20) // maxMemory 32MB
			if err != nil {
				log.Println("parse multipart form fail:", err)
			}
		}

		formDecoder := schema.NewDecoder()
		formDecoder.SetAliasTag("json")

		msg := Message{}
		if err := formDecoder.Decode(&msg, r.PostForm); err != nil {
			log.Println("unmarshal fail:", err)
			http.Error(w, "unmarshal fail", http.StatusInternalServerError)
			return
		}

		if err := fn(srv, msg); err != nil {
			log.Println("callback fail:", err)
			http.Error(w, "callback fail", http.StatusInternalServerError)
			return
		}

		log.Println("got message:", msg.Text)
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
