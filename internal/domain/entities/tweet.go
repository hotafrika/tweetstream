package entities

import "time"

// Tweet represents tweet object
type Tweet struct {
	Text      string    `json:"text"`
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Author    Author    `json:"author"`
}

// Author represents the creator of a tweet
type Author struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	URL      string `json:"url"`
	Location string `json:"location"`
}
