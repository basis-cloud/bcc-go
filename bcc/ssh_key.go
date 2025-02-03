package bcc

type SshKey struct {
	manager   *Manager
	ID        string `json:"id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func NewSshKey(name string, publicKey string) SshKey {
	k := SshKey{Name: name, PublicKey: publicKey}
	return k
}

func (m *Manager) GetSshKeys() (sshKeys []*SshKey, err error) {
	path := "v1/account/me/key"
	err = m.GetItems(path, Defaults(), &sshKeys)
	for i := range sshKeys {
		sshKeys[i].manager = m
	}
	return
}
