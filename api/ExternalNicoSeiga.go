/*
ExternalNicoSeiga.go
Yayoi

Created by Cure Gecko on 5/20/15.
Copyright 2015, Cure Gecko. All rights reserved.

Nico Seiga Data Puller.
*/
package main

import (
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type ExternalNicoSeiga struct {
	Server *Iori
}

func (e *ExternalNicoSeiga) Name() string {
	return "Nico Seiga"
}

func (e *ExternalNicoSeiga) Match(URL string) bool {
	if regexp.MustCompile("(?i).*seiga\\.nicovideo\\.jp/seiga/im[0-9]+").MatchString(URL) {
		return true
	} else if regexp.MustCompile("(?i).*seiga\\.nicovideo\\.jp/image/source/[0-9]+$").MatchString(URL) {
		return true
	} else if regexp.MustCompile("(?i).*lohas\\.nicoseiga\\.jp/(?:priv|o)/.*/[0-9]+$").MatchString(URL) {
		return true
	}
	return false
}

func (e *ExternalNicoSeiga) request(method, URL string, body io.Reader) (*http.Request, error) {
	session := e.Server.Settings.Get("NicoSession")
	if session == "" {
		form := url.Values{}
		form.Add("mail_tel", e.Server.Settings.Get("NicoUsername"))
		form.Add("password", e.Server.Settings.Get("NicoPassword"))

		req, _ := http.NewRequest("POST", "https://secure.nicovideo.jp/secure/login?show_button_twitter=1&site=nicoaccount&show_button_facebook=1&next_url=/my/account", strings.NewReader(form.Encode()))
		req.Header.Add("Referer", "https://secure.nicovideo.jp/")
		req.Header.Add("User-Agent", FirefoxUserAgent)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		transport := http.Transport{}
		resp, err := transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 && resp.StatusCode != 302 {
			return nil, newExternalError("Incorrect response " + resp.Status)
		}

		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "user_session" {
				session = cookie.Value
				e.Server.Settings.Set("NicoSession", session)
			}
		}
		resp.Body.Close()
		if session == "" {
			return nil, newExternalError("Unable to login to nico")
		}
	}

	req, _ := http.NewRequest(method, URL, body)
	req.Header.Add("User-Agent", FirefoxUserAgent)
	req.Header.Add("Cookie", "user_session="+session)
	return req, nil
}

func (e *ExternalNicoSeiga) Get(URL string) (*ExternalData, error) {
	_, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	var nicoID string
	matches := regexp.MustCompile("(?i).*seiga\\.nicovideo\\.jp/seiga/im([0-9]+)").FindStringSubmatch(URL)
	if len(matches) == 2 {
		nicoID = matches[1]
	}
	matches = regexp.MustCompile("(?i).*seiga\\.nicovideo\\.jp/image/source/([0-9]+)$").FindStringSubmatch(URL)
	if len(matches) == 2 {
		nicoID = matches[1]
	}
	matches = regexp.MustCompile("(?i).*lohas\\.nicoseiga\\.jp/(?:priv|o)/.*/([0-9]+)$").FindStringSubmatch(URL)
	if len(matches) == 2 {
		nicoID = matches[1]
	}
	nicoURL := "http://seiga.nicovideo.jp/seiga/im" + nicoID

	req, err := e.request("GET", nicoURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	found := doc.Find("a#link_btn_login").Length()
	if found == 1 {
		e.Server.Settings.Remove("NicoSession")
		req, err := e.request("GET", nicoURL, nil)
		if err != nil {
			return nil, err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		doc, err = goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return nil, err
		}
	}

	externalData := new(ExternalData)
	externalData.Post = nicoURL

	authorLink := doc.Find("div#ko_watchlist_info li.user_name a")
	authorURL := authorLink.AttrOr("href", "")
	if authorURL == "" {
		return nil, newExternalError("Unable to find author information.")
	}
	var authorID string
	matches = regexp.MustCompile("(?i)/user/illust/([0-9]+)$").FindStringSubmatch(authorURL)
	if len(matches) == 2 {
		authorID = matches[1]
	}
	externalData.AuthorURL = "http://seiga.nicovideo.jp/user/illust/" + authorID
	externalData.AuthorName = authorLink.Find("strong").Text()

	sourceURL := "http://seiga.nicovideo.jp/image/source/" + nicoID

	req, err = e.request("GET", sourceURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err = goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	originalLink := doc.Find("div.illust_view_big img").AttrOr("src", "")
	if originalLink == "" {
		return nil, newExternalError("Unable to find image link.")
	}
	externalData.Images = append(externalData.Images, "http://lohas.nicoseiga.jp"+originalLink)

	return externalData, nil
}
