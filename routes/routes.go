package routes

import (
	"net/http"
)

type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}
