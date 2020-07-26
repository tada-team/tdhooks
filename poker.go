package main

import (
	"fmt"
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

	if cfg.GotMessage == "" {
		cfg.GotMessage = "Got: %d"
	}

	stopRegex := regexp.MustCompile(`^[\s\-]+$`)

	type state struct{ votes []int }
	states := make(map[string]state, 0)

	listen(cfg.Path, rtr, servers, func(srv server, msg Message) error {
		v, ok := states[msg.Chat]
		if stopRegex.MatchString(msg.Text) {
			avg := 0
			n := len(v.votes)
			if n > 0 {
				for _, vote := range v.votes {
					avg += vote
				}
				avg = int(math.Ceil(float64(avg) / float64(n)))
			}
			states[msg.Chat] = state{votes: []int{}}
			if err := srv.SendMessage(fmt.Sprintf(cfg.VoteMessage, n, avg)); err != nil {
				return err
			}
		} else if vote := forceInt64(msg.Text); vote > 0 {
			if !ok {
				states[msg.Chat] = state{votes: []int{int(vote)}}
			} else {
				v.votes = append(v.votes, int(vote))
			}
			if err := srv.SendMessage(fmt.Sprintf(cfg.GotMessage, vote)); err != nil {
				return err
			}
		}
		return nil
	})
}
