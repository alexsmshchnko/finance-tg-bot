package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"finance-tg-bot/internal/entity"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DocProcessor interface {
	PostDocument(ctx context.Context, doc *entity.Document) (err error)
	DeleteDocument(ctx context.Context, doc *entity.Document) (err error)
	GetDocumentCategories(ctx context.Context, user_id int, limit string) (cats []entity.TransCatLimit, err error)
	EditCategory(ctx context.Context, cat *entity.TransCatLimit) (err error)
	GetDocumentSubCategories(ctx context.Context, user_id int, trans_cat string) (subcats []string, err error)
}

func (r *Repository) PostDocument(ctx context.Context, doc *entity.Document) (err error) {
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

func (r *Repository) DeleteDocument(ctx context.Context, doc *entity.Document) (err error) {
	if doc.MsgID == "" || doc.UserId == 0 {
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

func (r *Repository) GetDocumentCategories(ctx context.Context, user_id int, limit string) (cats []entity.TransCatLimit, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/category/%d", r.serviceDomain, user_id),
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
	cat.Category = strings.ToLower(strings.TrimSpace(cat.Category))
	//

	jsonStr, err := json.Marshal(cat)
	if err != nil {
		r.Logger.Error("Repository.EditCategory json.Marshal(cat)", "err", err)
		return
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/category/%d", r.serviceDomain, cat.UserId),
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

func (r *Repository) GetDocumentSubCategories(ctx context.Context, user_id int, trans_cat string) (subcats []string, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"OPTIONS",
		fmt.Sprintf("%s/category/%d/%s", r.serviceDomain, user_id, trans_cat),
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
