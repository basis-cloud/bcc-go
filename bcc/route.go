package bcc

import (
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
	path, err := url.JoinPath("v1/router", r.ID, "route", id)
	if err != nil {
		return
	}
	err = r.manager.Get(path, Defaults(), &route)
	route.router = r
	return
}

func (r *Router) CreateRoute(route *Route) (err error) {
	path, err := url.JoinPath("v1/router", r.ID, "route")
	if err != nil {
		return err
	}
	args := &struct {
		Destination string `json:"destination"`
		NextHop     string `json:"nexthop"`
	}{
		Destination: route.Destination,
		NextHop:     route.NextHop,
	}
	err = r.manager.Request(http.MethodPost, path, args, &route)
	route.router = r
	return
}

func (route *Route) Update() error {
	path, err := url.JoinPath("v1/router", route.router.ID, "route", route.ID)
	if err != nil {
		return err
	}
	args := &struct {
		Destination string `json:"destination"`
		NextHop     string `json:"nexthop"`
	}{
		Destination: route.Destination,
		NextHop:     route.NextHop,
	}
	err = route.router.manager.Request(http.MethodPut, path, args, &route)
	return err
}

func (route *Route) Delete() error {
	path, err := url.JoinPath("v1/router", route.router.ID, "route", route.ID)
	if err != nil {
		return err
	}
	return route.router.manager.Delete(path, Defaults(), nil)
}

func (route Route) WaitLock() (err error) {
	path, err := url.JoinPath("v1/router", route.router.ID, "route", route.ID)
	if err != nil {
		return err
	}
	return loopWaitLock(route.router.manager, path)
}
