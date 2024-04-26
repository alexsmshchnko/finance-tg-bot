package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	db   *Ydb
	once sync.Once
)

func Handler(rw http.ResponseWriter, req *http.Request) {
	var (
		err    error
		resVal []byte
	)
	ctx := context.Background()

	once.Do(func() {
		db, err = connectDB(ctx, os.Getenv("YDB_DSN"), "", os.Getenv("PREFIX"))
		if err != nil {
			rw.WriteHeader(http.StatusBadGateway)
			rw.Write([]byte(fmt.Sprintf("connectDB error, %v", err)))
			once = sync.Once{}
			return
		}
	})

	splitPath := strings.Split(req.URL.Path, "/")
	if len(splitPath) < 2 {
		return
	}
	switch splitPath[1] {
	case "document":
		doc := &DBDocument{}
		err = json.NewDecoder(req.Body).Decode(doc)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(fmt.Sprintf("json.NewDecoder(req.Body).Decode(doc) error, %v", err)))
			return
		}

		switch req.Method {
		case "POST":
			err = db.PostDocument(ctx, doc)
			if err != nil {
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte(fmt.Sprintf("db.PostDocument error, %v", err)))
				return
			}
		case "DELETE":
			err = db.DeleteDocument(ctx, doc)
			if err != nil {
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte(fmt.Sprintf("db.DeleteDocument error, %v", err)))
				return
			}
		}
	case "category":
		switch req.Method {
		case "GET":
			user_id, err := strconv.Atoi(splitPath[2])
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(fmt.Sprintf("strconv.Atoi error, %v", err)))
				return
			}
			res, err := db.GetDocumentCategories(ctx, user_id, "")
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(fmt.Sprintf("db.GetDocumentCategories error, %v", err)))
				return
			}
			resVal, err = json.Marshal(res)
			if err != nil {
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte(fmt.Sprintf("json.Marshal(res) error, %v", err)))
				return
			}
		case "POST":
			cat := &TransCatLimit{}
			err = json.NewDecoder(req.Body).Decode(cat)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(fmt.Sprintf("json.NewDecoder(req.Body).Decode(cat) error, %v", err)))
				return
			}
			err = db.EditCategory(ctx, cat)
			if err != nil {
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte(fmt.Sprintf("db.EditCategory error, %v", err)))
				return
			}
		case "OPTIONS":
			user_id, err := strconv.Atoi(splitPath[2])
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(fmt.Sprintf("strconv.Atoi error, %v", err)))
				return
			}
			res, err := db.GetDocumentSubCategories(ctx, user_id, splitPath[3])
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(fmt.Sprintf("db.GetDocumentSubCategories error, %v", err)))
				return
			}
			resVal, err = json.Marshal(res)
			if err != nil {
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte(fmt.Sprintf("json.Marshal(res) error, %v", err)))
				return
			}
		}
	}

	rw.Write(resVal)
}
