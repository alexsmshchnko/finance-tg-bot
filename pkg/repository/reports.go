package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"finance-tg-bot/internal/entity"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Repository struct {
	serviceURL *url.URL
	authHeader *http.Header
	*http.Client
	*slog.Logger
}

func NewRepository(ServiceDomain, AuthToken string, log *slog.Logger) (rep *Repository) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = transport.MaxIdleConns
	u, _ := url.Parse(ServiceDomain)
	rep = &Repository{
		serviceURL: u,
		authHeader: &http.Header{},
		Client:     &http.Client{Transport: transport, Timeout: 12 * time.Second},
		Logger:     log,
	}
	rep.authHeader.Add("Authorization", "Basic "+AuthToken)
	return
}

type Reporter interface {
	GetStatementCatTotals(ctx context.Context, p map[string]string) (rres []entity.ReportResult, err error)
	GetUserStats(ctx context.Context, user_id int) (stats entity.UserStats, err error)
}

func (r *Repository) GetStatementCatTotals(ctx context.Context, p map[string]string) (rres []entity.ReportResult, err error) {
	jsonStr, err := json.Marshal(p)
	if err != nil {
		r.Logger.Error("Repository.GetStatementCatTotals json.Marshal(p)", "err", err)
		return
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		r.serviceURL.JoinPath("report").String(),
		bytes.NewBuffer(jsonStr))
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

	err = json.NewDecoder(resp.Body).Decode(&rres)
	if err != nil {
		r.Logger.Error("json.NewDecoder(resp.Body).Decode(&rres)", "err", err)
	}

	return
}

func (r *Repository) GetUserStats(ctx context.Context, user_id int) (stats entity.UserStats, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		r.serviceURL.JoinPath("userstats", fmt.Sprint(user_id)).String(),
		nil,
	)
	if err != nil {
		r.Logger.Error("Repository.GetUserStats http.NewRequestWithContext", "err", err)
		return
	}

	req.Header = r.authHeader.Clone()
	resp, err := r.Client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.GetUserStats client.Do", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		r.Logger.Error("Repository.GetUserStats response", "StatusCode", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		r.Logger.Error("json.NewDecoder(resp.Body).Decode(&stats)", "err", err)
	}

	return
}
