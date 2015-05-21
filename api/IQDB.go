/*
IQDB.go
Yayoi

Created by Cure Gecko on 5/17/15.
Copyright 2015, Cure Gecko. All rights reserved.

IQDB Wrapper.
*/

package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type IQDBError struct {
	s string
}

func (e *IQDBError) Error() string {
	return e.s
}

func newIQDBError(s string) *IQDBError {
	err := new(IQDBError)
	err.s = s
	return err
}

type IQDBResult struct {
	Id     uint64
	Score  float64
	Width  int64
	Height int64
}

type IQDB struct {
	Debug bool
}

func initIQDB() (*IQDB, error) {
	iqdb := new(IQDB)
	iqdb.Debug = true
	conn, reader, err := iqdb.Connect()
	if err != nil {
		return nil, err
	}
	iqdb.ReadUntilReady(reader)

	fmt.Fprint(conn, "db_list\n")
	if iqdb.Debug {
		fmt.Println("> db_list")
	}

	response, err := iqdb.ReadUntilReady(reader)
	if err != nil {
		return nil, err
	}
	if len(response) == 0 {
		fmt.Fprint(conn, "load 0 simple "+FSAPIPath+"iqdb.db\n")
		response, err := iqdb.ReadUntilReady(reader)
		if err != nil {
			return nil, err
		}
		if len(response) == 0 {
			return nil, newIQDBError("Unable to load DB")
		}
	}
	conn.Close()

	return iqdb, nil
}

func (i *IQDB) Connect() (net.Conn, *bufio.Reader, error) {
	conn, err := net.Dial("tcp", iqdbAddress)
	if err != nil {
		return nil, nil, err
	}
	reader := bufio.NewReader(conn)
	return conn, reader, err
}

func (i *IQDB) ReadUntilReady(reader *bufio.Reader) ([]string, error) {
	var response []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return response, err
		}
		if i.Debug {
			fmt.Print("< " + line)
		}
		line = line[0 : len(line)-1]
		if line[0:3] == "000" { // Read message received
			break
		} else {
			response = append(response, line)
		}
	}
	return response, nil
}

func (i *IQDB) QueryImage(path string) ([]IQDBResult, int, error) {
	var results []IQDBResult
	conn, reader, err := i.Connect()
	if err != nil {
		return results, 0, err
	}

	i.ReadUntilReady(reader)

	cmd := fmt.Sprintf("query 0 0 5 %s\n", path)
	if i.Debug {
		fmt.Print("> " + cmd)
	}
	fmt.Fprint(conn, cmd)
	response, err := i.ReadUntilReady(reader)
	if err != nil {
		return results, 0, err
	}
	var matches int
	for _, r := range response {
		s := strings.Split(r, " ")
		if s[0] == "101" {
			i := strings.Split(s[1], "=")
			if i[0] == "matches" {
				matches, _ = strconv.Atoi(i[1])
			}
		} else if s[0] == "200" {
			id, _ := strconv.ParseUint(s[1], 10, 64)
			score, _ := strconv.ParseFloat(s[2], 64)
			width, _ := strconv.ParseInt(s[3], 10, 64)
			height, _ := strconv.ParseInt(s[4], 10, 64)
			result := IQDBResult{id, score, width, height}
			results = append(results, result)
		}
	}
	conn.Close()

	return results, matches, nil
}

func (i *IQDB) AddImage(id, width, height uint, path string) error {
	conn, reader, err := i.Connect()
	if err != nil {
		return err
	}

	i.ReadUntilReady(reader)

	cmd := fmt.Sprintf("add 0 %v %v %v:%s\n", id, width, height, path)
	if i.Debug {
		fmt.Print("> " + cmd)
	}
	fmt.Fprint(conn, cmd)
	response, err := i.ReadUntilReady(reader)
	if err != nil {
		return err
	}
	if len(response) == 0 {
		return newIQDBError("No images added.")
	}
	conn.Close()

	return nil
}
