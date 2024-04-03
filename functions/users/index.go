package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func Handler(rw http.ResponseWriter, req *http.Request) {
	var (
		err      error
		username string
		resVal   []byte
		userInfo *DBClient
	)
	username = strings.TrimPrefix(req.URL.Path, "/users/")

	if db == nil {
		fmt.Println("connectDB: new connection")
		db, err = connectDB(context.Background(), os.Getenv("YDB_DSN"), "")
		if err != nil {
			rw.WriteHeader(http.StatusBadGateway)
			rw.Write([]byte(fmt.Sprintf("connectDB error, %v", err)))
			return
		}
	} else {
		fmt.Println("connectDB: already connected")
	}

	userInfo, err = db.GetUserInfo(context.Background(), username)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(fmt.Sprintf("db.GetUserInfo error, %v", err)))
		return
	}
	resVal, err = json.Marshal(userInfo)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(fmt.Sprintf("json.Marshal(userInfo) error, %v", err)))
		return
	}
	io.WriteString(rw, string(resVal))
}
