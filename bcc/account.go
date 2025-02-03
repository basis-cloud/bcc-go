package bcc

type Account struct {
	manager  *Manager
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (m *Manager) GetAccount() (account *Account, err error) {
	path := "v1/account/me"
	err = m.Get(path, Defaults(), &account)
	if err != nil {
		return
	}
	account.manager = m
	return
}
