package models

import (
	"time"
)

type Label string

type User struct {
	Username  string    `json:"username"`
	Password  []byte
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	masterKey []byte
}

func NewUser(username, password, masterKey string) *User {
	now := time.Now()
	return &User{
		Username:  username,
		Password:  []byte(password),
		CreatedAt: now,
		UpdatedAt: now,
		masterKey: []byte(masterKey),
	}
}

func FromUser(user *User) User {
	return *user
}

func (user *User) IsMaster(computedMasterHash string) {
	return bytes.Equal([]byte(user.masterKey),[]byte(computedMasterHash))
}

