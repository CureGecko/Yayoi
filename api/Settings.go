/*
Settings.go
Yayoi

Created by Cure Gecko on 5/19/15.
Copyright 2015, Cure Gecko. All rights reserved.

Get/Set Settings.
*/
package main

import (
	"gopkg.in/gorp.v1"
)

type Settings struct {
	DBmap *gorp.DbMap
}

func (s *Settings) Get(key string) string {
	obj, err := s.DBmap.Get(Setting{}, key)
	if err != nil || obj == nil {
		return ""
	}
	setting := obj.(*Setting)
	return setting.Value
}

func (s *Settings) Set(key, value string) error {
	obj, err := s.DBmap.Get(Setting{}, key)
	if err != nil {
		return err
	}
	if obj == nil {
		setting := new(Setting)
		setting.Name = key
		setting.Value = value
		err = s.DBmap.Insert(setting)
	} else {
		setting := obj.(*Setting)
		setting.Value = value
		_, err = s.DBmap.Update(setting)
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *Settings) Remove(key string) error {
	obj, err := s.DBmap.Get(Setting{}, key)
	if err != nil {
		return err
	}
	if obj != nil {
		setting := obj.(*Setting)
		_, err = s.DBmap.Delete(setting)
	}
	if err != nil {
		return err
	}
	return nil
}
