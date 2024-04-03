package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"finance-tg-bot/internal/entity"
	"finance-tg-bot/pkg/ydb"
	"io"
	"log/slog"
	"net/http"
)

type Repository struct {
	ServiceDomain string
	AuthToken     string
	*ydb.Ydb
	*slog.Logger
}

type Reporter interface {
	GetStatementCatTotals(ctx context.Context, p map[string]string) (rres []entity.ReportResult, err error)
}

func (r *Repository) GetStatementCatTotals(ctx context.Context, p map[string]string) (rres []entity.ReportResult, err error) {
	jsonStr, err := json.Marshal(p)
	if err != nil {
		r.Logger.Error("Repository.GetStatementCatTotals json.Marshal(p)", "err", err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, "GET", r.ServiceDomain+"/report", bytes.NewBuffer(jsonStr))
	if err != nil {
		r.Logger.Error("Repository.GetStatementCatTotals http.NewRequest", "err", err)
		return
	}
	req.Header.Add("Authorization", "Basic "+r.AuthToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.GetStatementCatTotals client.Do", "err", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Logger.Error("Repository.GetStatementCatTotals io.ReadAll", "err", err)
		return
	}
	err = json.Unmarshal(body, &rres)
	if err != nil {
		r.Logger.Error("Repository.GetStatementCatTotals json.Unmarshal", "err", err)
	}

	return
}
