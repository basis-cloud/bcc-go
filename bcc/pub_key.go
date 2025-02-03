package bcc

import (
	"fmt"
)

type PubKey struct {
	manager     *Manager
	ID          string `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
}

func (m *Manager) GetPublicKeys(account_id string) (public_keys []*PubKey, err error) {
	path := fmt.Sprintf("/v1/account/%s/key", account_id)
	err = m.GetItems(path, Defaults(), &public_keys)
	for i := range public_keys {
		public_keys[i].manager = m
	}
	return
}

func (a *Account) GetPublicKeys() (public_keys []*PubKey, err error) {
	public_keys, err = a.manager.GetPublicKeys(a.ID)
	if err != nil {
		return nil, err
	}
	return public_keys, nil
}

func (m *Manager) GetPublicKey(id string) (pub_key *PubKey, err error) {
	account, err := m.GetAccount()
	path := fmt.Sprintf("/v1/account/%s/key/%s", account.ID, id)
	err = m.Get(path, Defaults(), &pub_key)
	if err != nil {
		return
	}
	pub_key.manager = m
	return
}
