package types

import (
	"time"
)

type Label string

type Data struct {
	Lname      string    `json:"lname"`
	Lpassword  []byte    `json:"lpassword"`
	Lurl       string    `json:"lurl"`
	LastAccess time.Time `json:"last_access"`
	Owner			 string		 `json:"owner"`
}