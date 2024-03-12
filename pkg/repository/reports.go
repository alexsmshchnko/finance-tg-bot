package repository

import (
	"context"
	"database/sql"
	"finance-tg-bot/internal/entity"
	"finance-tg-bot/pkg/postgres"
	"finance-tg-bot/pkg/ydb"
	"log/slog"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

type Repository struct {
	*postgres.Postgres
	*ydb.Ydb
}

type Reporter interface {
	GetStatementTotals(ctx context.Context, log *slog.Logger, p map[string]string) (res []entity.ReportResult, err error)
	GetStatementCatTotals(ctx context.Context, p map[string]string) (rres []entity.ReportResult, err error)
}

func (r *Repository) GetStatementCatTotals(ctx context.Context, p map[string]string) (rres []entity.ReportResult, err error) {
	query := `DECLARE $client_id   AS String;
	          DECLARE $datefrom    AS Datetime;
			  DECLARE $dateto      AS Datetime;
select trans_cat
      ,case direction
         when 0 then sum(trans_amount)
         else direction * sum(trans_amount) 
       end as t_sum
      ,sum(trans_amount) as ta, 1 as tp
  from document
 where trans_date between $datefrom and $dateto
   and client_id = $client_id
   and trans_amount != 0
 group by trans_cat, direction
union all
select case direction
         when -1 then 'Total debit'
         when 0  then 'Total deposit'
         else         'Total credit'
       end as trans_cat
      ,case direction
         when 0 then sum(trans_amount)
         else direction * sum(trans_amount) 
       end as t_sum
      ,sum(trans_amount) as ta, 2 as tp
  from document
 where trans_date between $datefrom and $dateto
   and client_id = $client_id
   and trans_amount != 0
 group by direction
 order by tp ASC, ta DESC;`

	err = r.Ydb.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		t1, _ := time.Parse("02.01.2006", p["datefrom"])
		t2, _ := time.Parse("02.01.2006", p["dateto"])
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			query,
			table.NewQueryParameters(
				table.ValueParam("$client_id", types.BytesValueFromString(p["username"])),
				table.ValueParam("$datefrom", types.DatetimeValueFromTime(t1)),
				table.ValueParam("$dateto", types.DatetimeValueFromTime(t2)),
			),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		if err = res.NextResultSetErr(ctx); err != nil {
			return err
		}
		var (
			rn sql.NullString
			rv sql.NullInt64
		)
		for res.NextRow() {
			err = res.ScanNamed(
				named.Optional("trans_cat", &rn),
				named.Optional("t_sum", &rv),
			)
			if err != nil {
				return err
			}
			rres = append(rres, entity.ReportResult{Name: rn.String, Val: int(rv.Int64)})
		}
		return res.Err() // for driver retry if not nil
	})
	return rres, err
}

func (r *Repository) GetStatementTotals(ctx context.Context, log *slog.Logger, p map[string]string) (res []entity.ReportResult, err error) {
	query := `
	select case grouping(trans_cat)
			 when 1 then 'Total '
					  || case direction
					       when 1 then 'credit'
						   when 0 then 'deposit'
						   else 'debit'
						 end
			 else trans_cat
		   end as trans_cat
		  ,case direction
		     when 0 then sum(trans_amount)
			 else direction * sum(trans_amount) 
		   end as t_sum
	  from document
	 where trans_date between to_date($1,'dd.mm.yyyy') and to_date($2,'dd.mm.yyyy')
	   and client_id = $3
	   and trans_amount != 0
	 group by grouping sets ((trans_cat, direction), (direction)) 
	 order by grouping(trans_cat) asc, sum(trans_amount) desc`

	rows, err := r.QueryContext(
		ctx,
		query,
		p["datefrom"],
		p["dateto"],
		p["username"],
	)

	if err != nil {
		log.Error("statement query builder", "err", err)
		return nil, err
	}
	defer rows.Close()

	row := &entity.ReportResult{}
	for rows.Next() {
		err := rows.Scan(&row.Name, &row.Val)
		if err != nil {
			log.Error("can't scan sql rows", "err", err)
			return nil, err
		}
		res = append(res, *row)
	}
	return res, err
}
