package repository

import (
	"context"
	"finance-tg-bot/internal/entity"
	"finance-tg-bot/pkg/postgres"
	"log/slog"
)

type Repository struct {
	*postgres.Postgres
}

type Reporter interface {
	GetStatementTotals(ctx context.Context, log *slog.Logger, p map[string]string) (res []entity.ReportResult, err error)
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
