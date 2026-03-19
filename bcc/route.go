package bcc

import (
	"log"
	"net/http"
	"net/url"
)

type Route struct {
	router      *Router
	ID          string `json:"id"`
	Destination string `json:"destination"`
	NextHop     string `json:"nexthop"`
}

func NewRoute(destination, nexthop string) Route {
	return Route{
		Destination: destination,
		NextHop:     nexthop,
	}
}

func (r *Router) GetRoute(id string) (route *Route, err error) {
	path, _ := url.JoinPath("v1/router", r.ID, "route", id)

	if err = r.manager.Get(path, Defaults(), &route); err != nil {
		log.Printf("[REQUEST-ERROR]: get-route was failed: %s", err)
	} else {
		route.router = r
	}

	return
}

func (r *Router) CreateRoute(route *Route) (err error) {
	path, _ := url.JoinPath("v1/router", r.ID, "route")
	args := &struct {
		Destination string `json:"destination"`
		NextHop     string `json:"nexthop"`
	}{
		Destination: route.Destination,
		NextHop:     route.NextHop,
	}

	if err = r.manager.Request(http.MethodPost, path, args, &route); err != nil {
		log.Printf("[REQUEST-ERROR]: create-route was failed: %s", err)
	} else {
		route.router = r
	}

	return
}

func (route *Route) Update() (err error) {
	path, _ := url.JoinPath("v1/router", route.router.ID, "route", route.ID)
	args := &struct {
		Destination string `json:"destination"`
		NextHop     string `json:"nexthop"`
	}{
		Destination: route.Destination,
		NextHop:     route.NextHop,
	}

	if err = route.router.manager.Request(http.MethodPut, path, args, &route); err != nil {
		log.Printf("[REQUEST-ERROR]: update-route was failed: %s", err)
	}

	return
}

func (route *Route) Delete() error {
	path, _ := url.JoinPath("v1/router", route.router.ID, "route", route.ID)
	return route.router.manager.Delete(path, Defaults(), nil)
}

func (route Route) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/router", route.router.ID, "route", route.ID)
	return loopWaitLock(route.router.manager, path)
}
