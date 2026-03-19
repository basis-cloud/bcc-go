package bcc

import (
	"log"
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
	path := "v1/client"
	args := Defaults()
	args.merge(extraArgs)

	if err = m.GetItems(path, args, &clients); err != nil {
		log.Printf("[REQUEST-ERROR] get-clients was failed: %s", err)
	} else {
		for i := range clients {
			clients[i].manager = m
		}
	}

	return
}

func (m *Manager) GetClient(id string) (client *Client, err error) {
	path, _ := url.JoinPath("v1/client", id)

	if err = m.Get(path, Defaults(), &client); err != nil {
		log.Printf("[REQUEST-ERROR] get-client with id='%s' was failed: %s", id, err)
	} else {
		client.manager = m
	}

	return
}
