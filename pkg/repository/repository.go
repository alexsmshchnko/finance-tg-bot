package repository

import (
	"context"
	"finance-tg-bot/pkg/entity"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
)

// type Repository struct {
// 	*postgres.Postgres
// }

type DocumentRepo interface {
	Save(ctx context.Context, log *slog.Logger, doc entity.DBDocument) (err error)
	Get(ctx context.Context, log *slog.Logger, filter string) (docs []entity.DBDocument, err error)
}

func (r *Repository) Get(ctx context.Context, log *slog.Logger, filter string) (docs []entity.DBDocument, err error) {
	return
}

func (r *Repository) Save(ctx context.Context, log *slog.Logger, doc entity.DBDocument) (err error) {
	if doc.Direction == 0 {
		log.Debug("input direction is zero")
		query, args, err := sq.Select("direction").From("trans_category").
			Where(sq.Eq{"client_id": doc.ClientID, "trans_cat": doc.Category, "active": true}).
			PlaceholderFormat(sq.Dollar).ToSql()

		if err != nil {
			log.Error("direction query builder", "err", err)
			return err
		}
		err = r.GetContext(ctx, &doc.Direction, query, args...)
		if err != nil {
			log.Error("direction query execution", "err", err)
			return err
		}
		log.Debug("got from trans_cat table", "direction", doc.Direction)
	}

	query, args, err := sq.Insert("document").
		Columns("trans_date", "trans_cat", "trans_amount", "comment", "tg_msg_id", "client_id", "direction").
		Values(doc.Time, doc.Category, doc.Amount, doc.Description, doc.MsgID, doc.ClientID, doc.Direction).Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		log.Error("insert query builder", "err", err)
		return err
	}

	tx, _ := r.BeginTx(ctx, nil)
	rows, err := tx.Query(query, args...)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			log.Error("rollback err", "err", err)
			return err
		}

		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&doc.ID); err != nil {
			log.Error("can't scan document.ID", "err", err)
			return err
		}
	}
	log.Debug("document inserted", "id", doc.ID)

	return tx.Commit()
}
