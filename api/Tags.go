/*
Tags.go
Yayoi

Created by Cure Gecko on 5/19/15.
Copyright 2015, Cure Gecko. All rights reserved.

Tags.
*/

package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

type Tags struct {
	Server  *Iori
	Auth    *Auth
	Writer  http.ResponseWriter
	Request *http.Request
	Path    []string
}

func (u Tags) Process() {
	type Info struct {
		Success bool
		Tags    []*Tag
	}
	info := Info{}

	tag := u.Request.Form.Get("tag")

	if strings.Index(tag, "%") != -1 {
		info.Success = false
	} else {
		results, err := u.Server.DBmap.Select(Tag{}, "SELECT * FROM tags WHERE `Value` LIKE ? ORDER BY `UseCount`,`Value` LIMIT 10", "%"+tag+"%")
		if err != nil {
			log.Println(err)
			info.Success = false
		} else {
			for _, result := range results {
				tag := result.(*Tag)
				info.Tags = append(info.Tags, tag)
			}
		}
	}

	t, _ := template.ParseFiles("resources/tags.html")
	t.Execute(u.Writer, info)
}
