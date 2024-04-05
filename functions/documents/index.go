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
		err    error
		resVal []byte
	)
	ctx := context.Background()

	if db == nil {
		fmt.Println("connectDB: new connection")
		db, err = connectDB(ctx, os.Getenv("YDB_DSN"), "")
		if err != nil {
			rw.WriteHeader(http.StatusBadGateway)
			rw.Write([]byte(fmt.Sprintf("connectDB error, %v", err)))
			return
		}
	} else {
		fmt.Println("connectDB: already connected")
	}

	splitPath := strings.Split(req.URL.Path, "/")

	switch splitPath[1] {
	case "document":
		d, err := io.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadGateway)
			rw.Write([]byte(fmt.Sprintf("io.ReadAll(req.Body) error, %v", err)))
			return
		}
		doc := &DBDocument{}
		err = json.Unmarshal(d, doc)
		if err != nil {
			rw.WriteHeader(http.StatusBadGateway)
			rw.Write([]byte(fmt.Sprintf("json.Unmarshal(d, cat) error, %v", err)))
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
			res, err := db.GetDocumentCategories(ctx, splitPath[2], "")
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
			d, err := io.ReadAll(req.Body)
			if err != nil {
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte(fmt.Sprintf("io.ReadAll(req.Body) error, %v", err)))
				return
			}
			cat := &TransCat{}
			err = json.Unmarshal(d, cat)
			if err != nil {
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte(fmt.Sprintf("json.Unmarshal(d, cat) error, %v", err)))
				return
			}
			err = db.EditCategory(ctx, cat)
			if err != nil {
				rw.WriteHeader(http.StatusBadGateway)
				rw.Write([]byte(fmt.Sprintf("db.EditCategory error, %v", err)))
				return
			}
		case "OPTIONS":
			res, err := db.GetDocumentSubCategories(ctx, splitPath[2], splitPath[3])
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

	io.WriteString(rw, string(resVal))
}
