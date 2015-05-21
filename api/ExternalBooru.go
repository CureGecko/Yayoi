/*
ExternalBooru.go
Yayoi

Created by Cure Gecko on 5/20/15.
Copyright 2015, Cure Gecko. All rights reserved.

Booru Data Puller.
*/
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

type ExternalBooru struct {
	Server *Iori
}

func (e *ExternalBooru) Name() string {
	return "Booru"
}

func (e *ExternalBooru) Match(URL string) bool {
	if regexp.MustCompile("(?i).*donmai\\.us(?:/post/show/|/posts/|/ssd/data/|/data/)").MatchString(URL) {
		return true
	} else if regexp.MustCompile("(?i).*(?:gelbooru\\.com|youhate\\.us|safebooru\\.org)(?:.*page=post.*|/images/|/samples/|/thumbnails/)").MatchString(URL) {
		return true
	} else if regexp.MustCompile("(?i).*(?:konachan\\.com|yande\\.re)(?:/post/show/|/image/|/data/|/sample/|/jpeg/)").MatchString(URL) {
		return true
	}
	return false
}

func (e *ExternalBooru) Get(URL string) (*ExternalData, error) {
	_, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Add("url", URL)
	resp, err := http.PostForm("https://cure.ninja/booru/api/json/url/", form)
	var bytes []byte
	if err != nil {
		return nil, err
	}
	bytes, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var response map[string]interface{}
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return nil, err
	}
	if !response["success"].(bool) || len(response["results"].([]interface{})) == 0 {
		return nil, newExternalError("No information found.")
	}

	externalData := new(ExternalData)
	externalData.Post = response["results"].([]interface{})[0].(map[string]interface{})["page"].(string)
	externalData.Images = append(externalData.Images, response["results"].([]interface{})[0].(map[string]interface{})["url"].(string))
	externalData.AuthorName = response["results"].([]interface{})[0].(map[string]interface{})["userName"].(string)
	externalData.AuthorURL = response["results"].([]interface{})[0].(map[string]interface{})["userURL"].(string)

	return externalData, nil
}
