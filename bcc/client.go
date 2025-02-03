package bcc

import (
	"net/url"
)

type Client struct {
	manager      *Manager
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	PaymentModel string  `json:"payment_model"`
	Balance      float32 `json:"contract.balance"`
}

func (m *Manager) GetClients(extraArgs ...Arguments) (clients []*Client, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/client"
	err = m.GetItems(path, args, &clients)
	for i := range clients {
		clients[i].manager = m
	}
	return
}

func (m *Manager) GetClient(id string) (client *Client, err error) {
	path, _ := url.JoinPath("v1/client", id)
	err = m.Get(path, Defaults(), &client)
	if err != nil {
		return
	}
	client.manager = m
	return
}
