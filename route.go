// Copyright 2012 Ernest Micklei. All rights reserved.
// Use of this source code is governed by a license 
// that can be found in the LICENSE file.

package restful

import (
	"net/http"
	"strings"
)

const RouteFunctionCalled = 0

// Signature of function that can be bound to a Route.
type RouteFunction func(*Request, *Response)

// Route binds a HTTP Method,Path,Consumes combination to a RouteFunction.
type Route struct {
	Method   string
	Produces string
	Consumes string
	Path     string
	Function RouteFunction

	pathParts []string
}

func (self *Route) postBuild() {
	self.pathParts = strings.Split(self.Path, "/")
}

// If the Route matches the request then handle it and return http.StatusOK.
// Return other appropriate http status values otherwise.
func (self *Route) dispatch(httpWriter http.ResponseWriter, httpRequest *http.Request) int {
	// the order of matching types are relevant
	matches, params := self.matchesPath(httpRequest.URL.Path)
	if !matches {
		return http.StatusNotFound
	}
	method := httpRequest.Method
	if self.Method != method {
		return http.StatusMethodNotAllowed
	}
	accept := httpRequest.Header.Get(HEADER_Accept)
	if !self.matchesAccept(accept) {
		return http.StatusUnsupportedMediaType
	}
	// check for POST and PUT only
	if (method == "POST") || (method == "PUT") {
		contentType := httpRequest.Header.Get(HEADER_ContentType)
		if !self.matchesContentType(contentType) {
			return http.StatusUnsupportedMediaType
		}
	}
	self.Function(&Request{httpRequest, params}, &Response{httpWriter, accept})
	return 0
}

// Return whether the mimeType matches what this Route can produce.
func (self Route) matchesAccept(mimeTypesWithQuality string) bool {
	// cheap test first
	if len(self.Produces) == 0 || strings.HasPrefix(self.Produces, "*/*") {
		return true
	}
	parts := strings.Split(mimeTypesWithQuality, ",")
	for _, each := range parts {
		withoutQuality := strings.Split(each, ";")[0]
		if strings.Index(self.Produces, withoutQuality) != -1 {
			return true
		}
	}
	return false
}

// Return whether the mimeType matches what this Route can consume.
func (self Route) matchesContentType(mimeType string) bool {
	// cheap test first
	if len(self.Consumes) == 0 || strings.HasPrefix(self.Consumes, "*/*") {
		return true
	}
	parts := strings.Split(mimeType, ",")
	for _, each := range parts {
		if strings.Index(self.Consumes, each) != -1 {
			return true
		}
	}
	return false
}

// Check if the URL path matches the parameterized path of the Route.
// If it does then return a map(s->s) with the values for each path parameter.
func (self Route) matchesPath(urlPath string) (bool, map[string]string) {
	urlParts := strings.Split(urlPath, "/")
	if len(self.pathParts) != len(urlParts) {
		return false, nil
	}
	pathParameters := map[string]string{}
	for i, key := range self.pathParts {
		value := urlParts[i]
		if strings.HasPrefix(key, "{") { // path-parameter
			pathParameters[strings.Trim(key, "{}")] = value
		} else { // fixed
			if key != value {
				return false, nil
			}
		}
	}
	return true, pathParameters
}