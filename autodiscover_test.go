// Copyright (c) 2014 Niklas Wolber

package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestParsing(t *testing.T) {
	req, err := parseRequest(strings.NewReader(
		`<?xml version="1.0" encoding="utf-8"?>
        <Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/mobilesync/requestschema/2006">
            <Request>
                <EMailAddress>me@example.com</EMailAddress>
                <AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/mobilesync/responseschema/2006</AcceptableResponseSchema>
            </Request>
        </Autodiscover>`))

	t.Log(req.AcceptableResponseSchema)
	t.Log(req.EMailAddress)
	if err != nil {
		t.Fatal(err.Error())
	}

	if req.AcceptableResponseSchema != "http://schemas.microsoft.com/exchange/autodiscover/mobilesync/responseschema/2006" {
		t.Fail()
	}

	if req.EMailAddress != "me@example.com" {
		t.Fail()
	}
}

type response2006 struct {
	XMLName      xml.Name `xml:"Autodiscover"`
	DisplayName  string   `xml:"Response>User>DisplayName"`
	EMailAddress string   `xml:"Response>User>EMailAddress"`
	URL          string   `xml:"Response>Action>Settings>Server>Url"`
	Name         string   `xml:"Response>Action>Settings>Server>Name"`
}

var bla = []struct {
	server string
	mail   string
}{
	{"mail.example.com", "me@example.com"},
	{"mail", "me@example.com"},
	{"example.com", "abc@123"},
}

func Test2006RequestProcessing_new(t *testing.T) {
	const (
		schema = "http://schemas.microsoft.com/exchange/autodiscover/mobilesync/responseschema/2006"
	)

	for _, tt := range bla {
		uri := fmt.Sprintf("https://%sMicrosoft-Server-ActiveSync", tt.server)

		s := NewService(tt.server, uri)
		rec := httptest.NewRecorder()
		s.processRequest(rec, request{
			AcceptableResponseSchema: schema,
			EMailAddress:             tt.mail,
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("http status want %d, got %d", http.StatusOK, rec.Code)
		}

		var res response2006
		decoder := xml.NewDecoder(rec.Body)
		if err := decoder.Decode(&res); err != nil {
			t.Fatal(err.Error())
		}

		// Schema is not checked, because unmarshalling nested attributes is a pain
		// See https://code.google.com/p/go/issues/detail?id=3633

		if res.DisplayName != tt.mail {
			t.Fatalf("want display name '%s', got '%s'", res.DisplayName, tt.mail)
		}

		if res.EMailAddress != tt.mail {
			t.Fatalf("want email address was '%s', got '%s'", res.EMailAddress, tt.mail)
		}

		if res.URL != uri {
			t.Fatalf("want url was '%s', got '%s'", res.URL, uri)
		}

		if res.Name != uri {
			t.Fatalf("want name was '%s', got '%s'", res.Name, uri)
		}
	}
}

type response2006a struct {
	XMLName     xml.Name                `xml:"Autodiscover"`
	DisplayName string                  `xml:"Response>User>DisplayName"`
	Protocols   []response2006aProtocol `xml:"Response>Protocol"`
}

type response2006aProtocol struct {
	XMLName   xml.Name `xml:"Protocol"`
	Server    string
	LoginName string
}

func Test2006aRequstProcessing(t *testing.T) {
	const (
		schema = "http://schemas.microsoft.com/exchange/autodiscover/mobilesync/responseschema/2006a"
	)

	for _, tt := range bla {
		uri := fmt.Sprintf("https://%sMicrosoft-Server-ActiveSync", tt.server)

		s := NewService(tt.server, uri)
		rec := httptest.NewRecorder()
		s.processRequest(rec, request{
			AcceptableResponseSchema: schema,
			EMailAddress:             tt.mail,
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("http status want %d, got %d", http.StatusOK, rec.Code)
		}

		var res response2006a
		decoder := xml.NewDecoder(rec.Body)
		if err := decoder.Decode(&res); err != nil {
			t.Fatal(err.Error())
		}

		if res.DisplayName != tt.mail {
			t.Fatalf("Response display name was '%s' expected '%s'", res.DisplayName, tt.mail)
		}

		if res.Protocols[0].Server != tt.server {
			t.Fatalf("Response server was '%s' expected '%s'", res.Protocols[0].Server, tt.server)
		}

		if res.Protocols[0].LoginName != tt.mail {
			t.Fatalf("Response server was '%s' expected '%s'", res.Protocols[0].LoginName, tt.mail)
		}

		if res.Protocols[1].Server != tt.server {
			t.Fatalf("Response server was '%s' expected '%s'", res.Protocols[1].Server, tt.server)
		}

		if res.Protocols[1].LoginName != tt.mail {
			t.Fatalf("Response server was '%s' expected '%s'", res.Protocols[1].LoginName, tt.mail)
		}
	}
}

func TestRenderResponse(t *testing.T) {

}

func TestRenderError(t *testing.T) {
	r := httptest.NewRecorder()
	renderError(r)

	if r.Code != http.StatusBadRequest {
		t.Fatalf("http status want %d, got %d", http.StatusBadRequest, r.Code)
	}
}
