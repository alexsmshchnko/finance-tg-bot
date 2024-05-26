package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type logger struct {
	*zap.Logger
	sync.Once
}

var (
	db  Ydb
	log logger
)

func Handler(rw http.ResponseWriter, req *http.Request) {
	var (
		resVal []byte
		err    error
	)
	ctx := context.Background()

	log.Once.Do(func() {
		config := zap.NewProductionConfig()
		// config.DisableCaller = true
		config.Level.SetLevel(zap.DebugLevel)
		log.Logger, err = config.Build()
		if err != nil {
			log.Once = sync.Once{}
			log.Error("config.Build err", zap.Error(err))
			return
		}
	})

	log.Info("new request to handle",
		zap.String("method", req.Method), zap.String("URL", req.URL.Path))

	db.Once.Do(func() {
		db.Driver, err = connectDB(ctx, os.Getenv("YDB_DSN"), "")
		if err != nil {
			db.Once = sync.Once{}
			rw.WriteHeader(http.StatusInternalServerError)
			log.Error("connectDB err", zap.Error(err))
			return
		}
		log.Info("db connected", zap.Any("name", db.Driver.Name()))
	})

	switch strings.Split(req.URL.Path, "/")[1] {
	case "report":
		p := &ReportParams{}
		err = json.NewDecoder(req.Body).Decode(p)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			log.Error("json.NewDecoder err", zap.Error(err))
			return
		}
		res, err := db.GetStatementCatTotals(ctx, p)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			log.Error("GetStatementCatTotals err", zap.Error(err))
			return
		}
		resVal, err = json.Marshal(res)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			log.Error("json.Marshal err", zap.Error(err))
			return
		}
	case "userstats":
		user_id, err := strconv.Atoi(strings.TrimPrefix(req.URL.Path, "/userstats/"))
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			log.Error("strconv.Atoi err", zap.Error(err))
			return
		}
		res, err := db.GetUserStats(ctx, user_id)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			log.Error("GetUserStats err", zap.Error(err))
			return
		}
		resVal, err = json.Marshal(res)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			log.Error("json.Marshal err", zap.Error(err))
			return
		}
	}
	rw.Write(resVal)
}
