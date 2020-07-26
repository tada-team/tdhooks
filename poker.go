package main

import (
	"fmt"
	"log"
	"math"
	"regexp"

	"github.com/gorilla/mux"
)

type pokerConfig struct {
	Path        string `yaml:"path"`
	VoteMessage string `yaml:"voteMessage"`
	GotMessage  string `yaml:"gotMessage"`
}

func (cfg pokerConfig) listen(rtr *mux.Router, servers []server) {
	if cfg.Path == "" {
		cfg.Path = "/poker"
	}

	if cfg.VoteMessage == "" {
		cfg.VoteMessage = "Num votes: %d, average: %d"
	}

	stopRegex := regexp.MustCompile(`^[\s\-]+$`)

	type state struct {
		votes map[string]int
	}

	states := make(map[string]*state)

	listen(cfg.Path, rtr, servers, func(srv server, msg Message) error {
		v, ok := states[srv.Key]
		if stopRegex.MatchString(msg.Text) {
			log.Println("poker: stop")
			avg := 0
			n := len(v.votes)
			if n > 0 {
				for _, vote := range v.votes {
					avg += vote
				}
				avg = int(math.Ceil(float64(avg) / float64(n)))
			}
			states[srv.Key] = &state{
				votes: make(map[string]int),
			}
			if err := srv.SendMessage(fmt.Sprintf(cfg.VoteMessage, n, avg)); err != nil {
				return err
			}
		} else if vote := forceInt64(msg.Text); vote > 0 {
			log.Println("poker: vote:", vote, "from:", msg.Sender)
			if !ok {
				states[srv.Key] = &state{
					votes: make(map[string]int),
				}
			}
			states[srv.Key].votes[msg.Sender] = int(vote)
			if cfg.GotMessage != "" {
				if err := srv.SendMessage(fmt.Sprintf(cfg.GotMessage, vote)); err != nil {
					return err
				}
			}
		} else {
			log.Println("poker: skip message")
		}
		return nil
	})
}
