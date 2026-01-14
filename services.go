package main

import "time"

type PollService interface {
	CreatePoll(req Poll) (int, error)
	GetPoll(id int64) (*Poll, error)
	ListPolls() []*Poll
	DeletePoll(id int64) error
	Vote(userEmail string, pollID int64, optionID int) (int, error)
}

type Poll struct {
	ID       int64            `json:"id"`
	Title    string           `json:"title"`
	Question string           `json:"question"`
	Options  []string         `json:"options"`
	Votes    map[int][]string `json:"votes"`
	Created  time.Time        `json:"created"`
}

type DatabasePollService struct {
}
