package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func Handler(rw http.ResponseWriter, req *http.Request) {
	var (
		err    error
		resVal []byte
	)

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

	userInfo, err := db.GetUserInfo(context.Background(), strings.TrimPrefix(req.URL.Path, "/users/"))
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
	rw.Write(resVal)
}
