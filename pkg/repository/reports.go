package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"finance-tg-bot/internal/entity"
	"io"
	"log/slog"
	"net/http"
)

type Repository struct {
	serviceDomain string
	authHeader    *http.Header
	*http.Client
	*slog.Logger
}

func NewRepository(ServiceDomain, AuthToken string, log *slog.Logger) (rep *Repository) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = transport.MaxIdleConns
	rep = &Repository{
		serviceDomain: ServiceDomain,
		authHeader:    &http.Header{},
		Client:        &http.Client{Transport: transport},
		Logger:        log,
	}
	rep.authHeader.Add("Authorization", "Basic "+AuthToken)
	return
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
	req, err := http.NewRequestWithContext(ctx, "GET", r.serviceDomain+"/report", bytes.NewBuffer(jsonStr))
	if err != nil {
		r.Logger.Error("Repository.GetStatementCatTotals http.NewRequest", "err", err)
		return
	}
	req.Header = r.authHeader.Clone()
	req.Header.Add("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
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
