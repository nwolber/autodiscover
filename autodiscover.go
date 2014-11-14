// Copyright (c) 2014 Niklas Wolber

package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"text/template"
	"time"
)

type request struct {
	XMLName                  xml.Name `xml:"Autodiscover"`
	AcceptableResponseSchema string   `xml:"Request>AcceptableResponseSchema"`
	EMailAddress             string   `xml:"Request>EMailAddress"`
}

const (
	schema2006  = "http://schemas.microsoft.com/exchange/autodiscover/mobilesync/responseschema/2006"
	schema2006a = "http://schemas.microsoft.com/exchange/autodiscover/mobilesync/responseschema/2006a"
)

var templates = template.Must(template.ParseFiles("2006.xml", "2006a.xml", "error.xml"))

func main() {

	const (
		defaultPort = 80
		usagePort   = "the port to listen on, default: %d"

		defaultServer = "localhost"
		usageServer   = "address of the mail server which should be returned"

		defaultServiceURI = "https://%s/Microsoft-Server-ActiveSync"
		usageServiceURI   = "the URI where the active sync server is located, where %s is the parameter given by -server"
	)

	var port int
	var server, URI string

	flag.IntVar(&port, "port", defaultPort, fmt.Sprintf(usagePort, defaultPort))
	flag.StringVar(&server, "server", "", usageServer)
	flag.StringVar(&URI, "URI", defaultServiceURI, usageServiceURI)
	flag.Parse()

	sc := NewService(server, URI)

	if l, err := net.Listen("tcp", ":"+string(port)); err != nil {
		fmt.Println(err)
	} else {
		sc.start(l)
	}

	// sc.start(port)
}

func (s *Service) start(l net.Listener) {
	http.Handle("/", s)

	if err := http.Serve(l, nil); err != nil {
		fmt.Println("Cannot start server")
		fmt.Println(err)
	} else {
		fmt.Println("Server started")
	}
}

// Service represents the state of the autodiscover service
type Service struct {
	server     string
	serviceURI string
}

// NewService initializes a new service instance
func NewService(server, serviceURI string) *Service {
	return &Service{server: server, serviceURI: serviceURI}
}

// ServeHTTP handles requests to the autodiscovering service
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if req, err := parseRequest(r.Body); err != nil {
		renderError(w)
	} else {
		s.processRequest(w, req)
	}
}

func parseRequest(r io.Reader) (req request, err error) {
	err = xml.NewDecoder(r).Decode(&req)
	return
}

func (s *Service) processRequest(w io.Writer, req request) {
	if req.AcceptableResponseSchema == schema2006 {
		renderResponse(w, "2006.xml", req.EMailAddress, req.EMailAddress, s.serviceURI)
	} else if req.AcceptableResponseSchema == schema2006a {
		renderResponse(w, "2006a.xml", req.EMailAddress, req.EMailAddress, s.server)
	}
}

func renderResponse(w io.Writer, tmpl, displayName, loginName, serviceURI string) {
	templates.ExecuteTemplate(w, tmpl, struct {
		DisplayName string
		LoginName   string
		ServiceURI  string
	}{
		displayName,
		loginName,
		serviceURI,
	})
}

func renderError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	templates.ExecuteTemplate(w, "error.xml", struct {
		Now time.Time
	}{
		time.Now(),
	})
}
