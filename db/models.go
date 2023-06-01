package db

import (
	"time"
)

type Account struct {
	Id        int64     `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"password"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	Session   string    `json:"session"`
}

type File struct {
	Id          int64     `json:"id"`
	Owner       int64     `json:"owner"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	LocationWav string    `json:"location_wav"`
	LocationTxt string    `json:"location_txt"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
	Guid        string    `json:"guid"`
}
