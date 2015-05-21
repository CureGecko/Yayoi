/*
ExternalPixiv.go
Yayoi

Created by Cure Gecko on 5/19/15.
Copyright 2015, Cure Gecko. All rights reserved.

Pixiv Data Puller.
*/
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const PixivClientID = "bYGKuGVw91e0NMfPGp44euvGt59s"
const PixivSecret = "HP3RmkgAmEGro0gn1x9ioawQE8WMfvLXDz3ZqxpK"

type ExternalPixiv struct {
	Server *Iori
}

func (e *ExternalPixiv) Name() string {
	return "Pixiv"
}

func (e *ExternalPixiv) Match(URL string) bool {
	if regexp.MustCompile("(?i).*pixiv.net/.*illust_id.*").MatchString(URL) {
		return true
	} else if regexp.MustCompile("(?i).*pixiv\\..*/works/[0-9]+.*").MatchString(URL) {
		return true
	} else if regexp.MustCompile("(?i).*i.*\\.pixiv\\.net/.*").MatchString(URL) && !regexp.MustCompile("(?i)/profile/").MatchString(URL) && !regexp.MustCompile("(?i)dic.pixiv.net").MatchString(URL) {
		return true
	}
	return false
}

func (e *ExternalPixiv) authIfNeeded() error {
	needAuth := false
	if e.Server.Settings.Get("PixivBearer") == "" {
		needAuth = true
	}

	expire, _ := strconv.ParseInt(e.Server.Settings.Get("PixivExpire"), 10, 64)

	if expire-time.Now().Unix() <= 0 {
		needAuth = true
	}

	if needAuth {
		form := url.Values{}
		if refresh := e.Server.Settings.Get("PixivRefresh"); refresh != "" {
			form.Add("refresh_token", refresh)
			form.Add("grant_type", "refresh_token")
		} else {
			form.Add("username", e.Server.Settings.Get("PixivUsername"))
			form.Add("password", e.Server.Settings.Get("PixivPassword"))
			form.Add("grant_type", "password")
		}
		form.Add("client_id", PixivClientID)
		form.Add("client_secret", PixivSecret)

		req, _ := http.NewRequest("POST", "https://oauth.secure.pixiv.net/auth/token", strings.NewReader(form.Encode()))
		req.Header.Add("Referer", "http://www.pixiv.net/")
		req.Header.Add("User-Agent", "PixivIOSApp/5.1.1")
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		bytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		var response map[string]interface{}
		err = json.Unmarshal(bytes, &response)
		if err != nil {
			return err
		}
		if _, ok := response["has_error"]; ok {
			return newExternalError("Request returned error.")
		}
		if _, ok := response["response"]; !ok {
			return newExternalError("No response exists.")
		}
		response = response["response"].(map[string]interface{})
		accessToken := response["access_token"].(string)
		refreshToken := response["refresh_token"].(string)
		expiresIn := int64(response["expires_in"].(float64))
		expires := time.Now().Unix() + (expiresIn - 20)

		e.Server.Settings.Set("PixivBearer", accessToken)
		e.Server.Settings.Set("PixivRefresh", refreshToken)
		e.Server.Settings.Set("PixivExpire", strconv.FormatInt(expires, 10))

		cookies := resp.Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "PHPSESSID" {
				e.Server.Settings.Set("PixivSession", cookie.Value)
			}
		}
	}
	return nil
}

func (e *ExternalPixiv) getOriginalURL(image string) string {
	u, err := url.Parse(image)
	if err != nil {
		return ""
	}
	path := strings.Split(u.Path, "/")
	var newPath []string
	lastPath := len(path) - 1
	foundImg := false
	for i, p := range path {
		if !foundImg && len(p) >= 3 && p[0:3] == "img" {
			foundImg = true
			newPath = append(newPath, "img-original")
		} else if foundImg && i != lastPath {
			newPath = append(newPath, p)
		} else if foundImg && i == lastPath {
			extension := p[strings.LastIndex(p, "."):]
			end := p[strings.LastIndex(p, "_"):]
			if end[0:2] != "_p" {
				p = p[0:strings.LastIndex(p, "_")] + extension
			}
			newPath = append(newPath, p)
		}
	}
	u.Path = strings.Join(newPath, "/")
	return u.String()
}

func (e *ExternalPixiv) findBestImage(images map[string]interface{}) string {
	var size int
	var thisImage string
	for key, value := range images {
		var thisSize int
		switch key {
		case "px_16x16":
			thisSize = 1
		case "px_48x48":
			thisSize = 2
		case "px_50x50":
			thisSize = 3
		case "px_56x56":
			thisSize = 4
		case "px_64x64":
			thisSize = 5
		case "px_120x":
			thisSize = 6
		case "px_128x128":
			thisSize = 7
		case "max_240x240":
			thisSize = 8
		case "ugoira600x600":
			thisSize = 9
		case "small":
			thisSize = 10
		case "medium":
			thisSize = 11
		case "ugoira1920x1080":
			thisSize = 12
		case "large":
			thisSize = 13
		default:
			log.Println("Unknown pixiv image size ", thisImage)
		}
		if size < thisSize {
			thisImage = value.(string)
		}
	}

	return e.getOriginalURL(thisImage)
}

func (e *ExternalPixiv) Get(URL string) (*ExternalData, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	var pixivID string
	if regexp.MustCompile("(?i).*pixiv.net/.*illust_id.*").MatchString(URL) {
		pixivID = u.Query().Get("illust_id")
	} else if regexp.MustCompile("(?i).*pixiv\\..*/works/[0-9]+.*").MatchString(URL) {
		path := strings.Split(u.Path, "/")
		if len(path) >= 3 {
			postID := path[2]
			if strings.Index(postID, ".") != -1 {
				postID = postID[0:strings.Index(postID, ".")]
			}
			pixivID = postID
		}
	} else if regexp.MustCompile("(?i).*i.*\\.pixiv\\.net/.*").MatchString(URL) && !regexp.MustCompile("(?i)/profile/").MatchString(URL) && !regexp.MustCompile("(?i)dic.pixiv.net").MatchString(URL) {
		path := strings.Split(u.Path, "/")
		postID := path[len(path)-1]
		postID = postID[0:strings.Index(postID, ".")]
		if strings.Index(postID, "_") != -1 {
			postID = postID[0:strings.Index(postID, "_")]
		}
		pixivID = postID
	}
	id, err := strconv.ParseInt(pixivID, 10, 64)
	if err != nil {
		return nil, err
	}

	APIURL := "https://public-api.secure.pixiv.net/v1/works/" + strconv.FormatInt(id, 10) + ".json"

	err = e.authIfNeeded()
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", APIURL, nil)
	req.Header.Add("Referer", "http://www.pixiv.net/")
	req.Header.Add("User-Agent", "PixivIOSApp/5.1.1")
	req.Header.Add("Authorization", "Bearer "+e.Server.Settings.Get("PixivBearer"))
	req.Header.Add("Cookie", "PHPSESSID="+e.Server.Settings.Get("PixivSession"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	if _, ok := data["has_error"]; ok {
		e.Server.Settings.Remove("PixivBearer")
		err = e.authIfNeeded()
		if err != nil {
			return nil, err
		}

		req, _ := http.NewRequest("GET", APIURL, nil)
		req.Header.Add("Referer", "http://www.pixiv.net/")
		req.Header.Add("User-Agent", "PixivIOSApp/5.1.1")
		req.Header.Add("Authorization", "Bearer "+e.Server.Settings.Get("PixivBearer"))
		req.Header.Add("Cookie", "PHPSESSID="+e.Server.Settings.Get("PixivSession"))
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		bytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var newData map[string]interface{}
		err = json.Unmarshal(bytes, &newData)
		if err != nil {
			return nil, err
		}
		data = newData

		if _, ok := data["has_error"]; ok {
			return nil, newExternalError("Work returned error.")
		}
	}
	if _, ok := data["response"]; !ok {
		return nil, newExternalError("No work exists.")
	}

	externalData := new(ExternalData)

	responses := data["response"].([]interface{})
	for _, response := range responses {
		post, ok := response.(map[string]interface{})
		if !ok {
			return nil, newExternalError("Response is incorrect.")
		}
		if post["is_manga"].(bool) {
			metadata, ok := post["metadata"].(map[string]interface{})
			if !ok {
				return nil, newExternalError("Should be metadata, but there isn't.")
			}
			pages, ok := metadata["pages"].([]interface{})
			if !ok {
				return nil, newExternalError("No pages.")
			}
			for _, page := range pages {
				pageInfo, ok := page.(map[string]interface{})
				if !ok {
					return nil, newExternalError("Page is incorrect.")
				}
				images, ok := pageInfo["image_urls"].(map[string]interface{})
				if !ok {
					return nil, newExternalError("Cannot get image urls.")
				}
				thisImage := e.findBestImage(images)
				externalData.Images = append(externalData.Images, thisImage)
			}
		} else {
			images, ok := post["image_urls"].(map[string]interface{})
			if !ok {
				return nil, newExternalError("Cannot get image urls.")
			}
			thisImage := e.findBestImage(images)
			externalData.Images = append(externalData.Images, thisImage)
		}
		externalData.Post = "http://www.pixiv.net/member_illust.php?mode=medium&illust_id=" + strconv.FormatInt(id, 10)
		user, ok := post["user"].(map[string]interface{})
		if !ok {
			return nil, newExternalError("Cannot get user info.")
		}
		name, ok := user["name"].(string)
		if !ok {
			return nil, newExternalError("Cannot get user name.")
		}
		externalData.AuthorName = name
		userID, ok := user["id"].(float64)
		if !ok {
			return nil, newExternalError("Cannot get user id.")
		}
		externalData.AuthorURL = "http://www.pixiv.net/member.php?id=" + strconv.FormatInt(int64(userID), 10)
	}

	return externalData, nil
}
