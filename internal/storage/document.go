package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

type DocumentPostgresStorage struct {
	connString string
}

func NewDocumentStorage(connString string) *DocumentPostgresStorage {
	return &DocumentPostgresStorage{connString: connString}
}

type dbDocument struct {
	ID          int64     `db:id`
	Time        time.Time `db:trans_date`
	Category    string    `db:trans_cat`
	Amount      int       `db:trans_amount`
	Description string    `db:comment`
	MsgID       string    `db:tg_msg_id`
	ClientID    string    `db:client_id`
}

func (s *DocumentPostgresStorage) GetCategories(ctx context.Context, username string) (cat []string, err error) {
	dbpool, err := pgxpool.New(ctx, s.connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		return
	}
	defer dbpool.Close()

	rows, err := dbpool.Query(ctx, "select cat from public.doc where client_id = $1 group by cat order by count(*) desc", username)
	var qRes []any
	for rows.Next() {
		rtrn, _ := rows.Values()
		qRes = append(qRes, rtrn...)

	}
	for i := 0; i < len(qRes); i++ {
		cat = append(cat, (qRes[i].(string)))
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return
	}

	return

}

func (s *DocumentPostgresStorage) SayHello(ctx context.Context, username string) (err error) {
	dbpool, err := pgxpool.New(ctx, s.connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		return err
	}
	defer dbpool.Close()

	var greeting string
	err = dbpool.QueryRow(ctx, "select 'Hello, "+username+"!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return err
	}

	fmt.Println(greeting)
	return
}
