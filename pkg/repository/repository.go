package repository

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"finance-tg-bot/internal/entity"
	"io"
	"net/http"
	"strings"
)

type DocProcessor interface {
	PostDocument(ctx context.Context, doc *DBDocument) (err error)
	DeleteDocument(ctx context.Context, doc *DBDocument) (err error)
	GetDocumentCategories(ctx context.Context, username, limit string) (cats []entity.TransCatLimit, err error)
	EditCategory(ctx context.Context, cat *entity.TransCatLimit) (err error)
	GetDocumentSubCategories(ctx context.Context, username, trans_cat string) (subcats []string, err error)
}

type DBDocument struct {
	RecDate     sql.NullTime   `json:"rec_time"`
	TransDate   sql.NullTime   `json:"trans_date"`
	Category    sql.NullString `json:"trans_cat"`
	Amount      sql.NullInt64  `json:"trans_amount"`
	Description sql.NullString `json:"comment"`
	MsgID       sql.NullString `json:"msg_id"`
	ChatID      sql.NullString `json:"chat_id"`
	ClientID    sql.NullString `json:"client_id"`
	Direction   sql.NullInt16  `json:"direction"`
}

func (r *Repository) PostDocument(ctx context.Context, doc *DBDocument) (err error) {
	//preformat
	doc.Description.String = strings.ToLower(strings.TrimSpace(doc.Description.String))
	//

	jsonStr, err := json.Marshal(doc)
	if err != nil {
		r.Logger.Error("Repository.PostDocument json.Marshal(doc)", "err", err)
		return
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		r.serviceDomain+"/document",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		r.Logger.Error("Repository.PostDocument http.NewRequestWithContext", "err", err)
		return
	}

	req.Header = r.authHeader.Clone()
	req.Header.Add("Content-Type", "application/json")
	resp, err := r.Client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.PostDocument client.Do", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		r.Logger.Error("Repository.PostDocument response", "StatusCode", resp.StatusCode)
		return
	}

	return
}

func (r *Repository) DeleteDocument(ctx context.Context, doc *DBDocument) (err error) {
	if !doc.MsgID.Valid || !doc.ClientID.Valid {
		r.Logger.Error("Repository.DeleteDocument not enough input params to delete document")
		return
	}

	jsonStr, err := json.Marshal(doc)
	if err != nil {
		r.Logger.Error("Repository.DeleteDocument json.Marshal(doc)", "err", err)
		return
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		r.serviceDomain+"/document",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		r.Logger.Error("Repository.DeleteDocument http.NewRequestWithContext", "err", err)
		return
	}

	req.Header = r.authHeader.Clone()
	req.Header.Add("Content-Type", "application/json")
	resp, err := r.Client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.DeleteDocument client.Do", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		r.Logger.Error("Repository.DeleteDocument response", "StatusCode", resp.StatusCode)
		return
	}

	return
}

func (r *Repository) GetDocumentCategories(ctx context.Context, username, limit string) (cats []entity.TransCatLimit, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		r.serviceDomain+"/category/"+username,
		nil,
	)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentCategories http.NewRequestWithContext", "err", err)
		return
	}

	req.Header = r.authHeader.Clone()
	resp, err := r.Client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentCategories client.Do", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		r.Logger.Error("Repository.GetDocumentCategories response", "StatusCode", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentCategories io.ReadAll", "err", err)
		return
	}
	err = json.Unmarshal(body, &cats)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentCategories json.Unmarshal", "err", err)
	}

	return
}

func (r *Repository) EditCategory(ctx context.Context, cat *entity.TransCatLimit) (err error) {
	//preformat
	cat.Category.String = strings.ToLower(strings.TrimSpace(cat.Category.String))
	//

	jsonStr, err := json.Marshal(cat)
	if err != nil {
		r.Logger.Error("Repository.EditCategory json.Marshal(cat)", "err", err)
		return
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		r.serviceDomain+"/category/"+cat.ClientID.String,
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		r.Logger.Error("Repository.EditCategory http.NewRequestWithContext", "err", err)
		return
	}

	req.Header = r.authHeader.Clone()
	req.Header.Add("Content-Type", "application/json")
	resp, err := r.Client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.EditCategory client.Do", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		r.Logger.Error("Repository.EditCategory response", "StatusCode", resp.StatusCode)
		return
	}

	return
}

func (r *Repository) GetDocumentSubCategories(ctx context.Context, username, trans_cat string) (subcats []string, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"OPTIONS",
		r.serviceDomain+"/category/"+username+"/"+trans_cat,
		nil,
	)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentSubCategories http.NewRequestWithContext", "err", err)
		return
	}

	req.Header = r.authHeader.Clone()
	resp, err := r.Client.Do(req)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentSubCategories client.Do", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New(http.StatusText(resp.StatusCode))
		r.Logger.Error("Repository.GetDocumentSubCategories response", "StatusCode", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentSubCategories io.ReadAll", "err", err)
		return
	}
	err = json.Unmarshal(body, &subcats)
	if err != nil {
		r.Logger.Error("Repository.GetDocumentSubCategories json.Unmarshal", "err", err)
	}

	return
}
