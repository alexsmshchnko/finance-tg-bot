package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func Handler(rw http.ResponseWriter, req *http.Request) {
	var (
		resVal []byte
		err    error
	)

	if db == nil {
		fmt.Println("connectDB: new connection")
		db, err = connectDB(context.Background(), os.Getenv("YDB_DSN"), "")
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("connectDB error, %v", err)))
			return
		}
	} else {
		fmt.Println("connectDB: already connected")
	}

	switch strings.Split(req.URL.Path, "/")[1] {
	case "report":
		r, err := io.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(fmt.Sprintf("io.ReadAll(req.Body) error, %v", err)))
			return
		}
		p := make(map[string]string)
		if err = json.Unmarshal(r, &p); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(fmt.Sprintf("json.Unmarshal(r, &p) error, %v", err)))
			return
		}
		res, err := db.GetStatementCatTotals(context.Background(), p)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("GetStatementCatTotals error, %v", err)))
			return
		}
		resVal, err = json.Marshal(res)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("json.Marshal(res) error, %v", err)))
			return
		}
		rw.Write(resVal)
	case "userstats":
		user_id, err := strconv.Atoi(strings.TrimPrefix(req.URL.Path, "/userstats/"))
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(fmt.Sprintf("strconv.Atoi error, %v", err)))
			return
		}
		res, err := db.GetUserStats(context.Background(), user_id)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("GetUserStats error, %v", err)))
			return
		}
		resVal, err = json.Marshal(res)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("json.Marshal(res) error, %v", err)))
			return
		}
		rw.Write(resVal)
	}
}
