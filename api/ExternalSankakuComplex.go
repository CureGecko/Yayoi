/*
ExternalSankakuComplex.go
Yayoi

Created by Cure Gecko on 5/20/15.
Copyright 2015, Cure Gecko. All rights reserved.

Sankaku Complex Data Puller.
*/
package main

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type ExternalSankakuComplex struct {
	Server *Iori
}

func (e *ExternalSankakuComplex) Name() string {
	return "Sankaku Complex"
}

func (e *ExternalSankakuComplex) Match(URL string) bool {
	if regexp.MustCompile("(?i)chan.sankakucomplex.com/post/show/[0-9]+").MatchString(URL) {
		return true
	} else if regexp.MustCompile("(?i)sankakucomplex.com(?:/image/|/data/|/sample/|/jpeg/)").MatchString(URL) {
		return true
	}
	return false
}

func (e *ExternalSankakuComplex) getID(MD5 string) string {
	resp, err := http.Get("https://chan.sankakucomplex.com/?tags=md5:" + MD5)
	if err != nil {
		return ""
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return ""
	}

	link := doc.Find("div.content span.thumb a")
	href, ok := link.Attr("href")
	if !ok {
		return ""
	}
	path := strings.Split(href, "/")
	id := path[len(path)-1]
	return id
}

func (e *ExternalSankakuComplex) Get(URL string) (*ExternalData, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	var sanID string
	if regexp.MustCompile("(?i)chan.sankakucomplex.com/post/show/[0-9]+").MatchString(URL) {
		path := strings.Split(u.Path, "/")
		sanID = path[len(path)-1]
	} else if regexp.MustCompile("(?i)sankakucomplex.com(?:/image/|/data/|/sample/|/jpeg/)").MatchString(URL) {
		path := strings.Split(u.Path, "/")
		fileName := path[len(path)-1]
		MD5 := fileName[0:strings.Index(fileName, ".")]
		sanID = e.getID(MD5)
	}
	if sanID == "" {
		return nil, newExternalError("Unable to get ID")
	}
	sanURL := "https://chan.sankakucomplex.com/post/show/" + sanID

	resp, err := http.Get(sanURL)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	externalData := new(ExternalData)
	externalData.Post = sanURL
	var userID string = "1"
	var userName string = "System"
	var image string
	doc.Find("div#stats li a").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if strings.Contains(href, "/user/show/") {
			path := strings.Split(href, "/")
			userID = path[len(path)-1]
			userName = s.Text()
		} else if strings.Contains(href, "cs.sankakucomplex.com/data/") {
			if href[0:2] == "//" {
				href = "https" + href
			}
			image = href
		}
	})
	externalData.AuthorURL = "https://chan.sankakucomplex.com/user/show/" + userID
	externalData.AuthorName = userName
	externalData.Images = append(externalData.Images, image)

	return externalData, nil
}
