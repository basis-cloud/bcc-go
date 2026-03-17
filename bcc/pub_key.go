package bcc

import (
	"fmt"
	"log"
)

type PubKey struct {
	manager     *Manager
	ID          string `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
}

func (m *Manager) GetPublicKeys(accountId string) (publicKeys []*PubKey, err error) {
	path := fmt.Sprintf("/v1/account/%s/key", accountId)

	if err = m.GetItems(path, Defaults(), &publicKeys); err != nil {
		log.Printf("[REQUEST-ERROR] get-public-keys was failed: %s", err)
	} else {
		for i := range publicKeys {
			publicKeys[i].manager = m
		}
	}

	return
}

func (a *Account) GetPublicKeys() (publicKeys []*PubKey, err error) {
	publicKeys, err = a.manager.GetPublicKeys(a.ID)
	return
}

func (m *Manager) GetPublicKey(id string) (publicKey *PubKey, err error) {
	account, err := m.GetAccount()
	if err != nil {
		log.Printf("[REQUEST-ERROR] get-public-key was failed: %s", err)
		return
	}
	path := fmt.Sprintf("/v1/account/%s/key/%s", account.ID, id)

	if err = m.Get(path, Defaults(), &publicKey); err != nil {
		log.Printf("[REQUEST-ERROR] get-public-key was failed: %s", err)
	} else {
		publicKey.manager = m
	}

	return
}
