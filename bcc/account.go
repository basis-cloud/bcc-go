package bcc

import "log"

type Account struct {
	manager  *Manager
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (m *Manager) GetAccount() (account *Account, err error) {
	path := "v1/account/me"

	if err = m.Get(path, Defaults(), &account); err != nil {
		log.Printf("[REQUEST-ERROR] get-account was failed: %s", err)
	} else {
		account.manager = m
	}

	return
}
