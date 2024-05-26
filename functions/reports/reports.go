package main

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	yc "github.com/ydb-platform/ydb-go-yc"
)

type Ydb struct {
	*ydb.Driver
	sync.Once
}

const (
	REGULAR_TOTALS = iota
	SUBCAT_TOTALS
	HISTORY_BY_CAT
	HISTORY_BY_SUBCAT
)

func connectDB(ctx context.Context, dsn, saPath string) (*ydb.Driver, error) {
	var opt ydb.Option
	if saPath == "" {
		// auth inside cloud (virual machine or yandex function)
		opt = yc.WithMetadataCredentials()
	} else {
		// auth from service account key file
		opt = yc.WithServiceAccountKeyFileCredentials(saPath)
	}
	return ydb.Open(ctx, dsn, yc.WithInternalCA(), opt)
}

type ReportParams struct {
	User_id     uint64
	Report_type string
	Date_from   time.Time
	Date_to     time.Time
	Add_attr1   string
	Add_attr2   string
}

type ReportResult struct {
	Name string `json:"name"`
	Val  int    `json:"val"`
}

func (r *Ydb) GetStatementCatTotals(ctx context.Context, p *ReportParams) (rres []ReportResult, err error) {
	var repType int
	switch p.Report_type {
	case "subcat":
		repType = SUBCAT_TOTALS
	case "historyCat":
		repType = HISTORY_BY_CAT
	case "historySubcat":
		repType = HISTORY_BY_SUBCAT
	default:
		repType = REGULAR_TOTALS
	}

	query := `DECLARE $client_id   AS Uint64;
	          DECLARE $datefrom    AS Datetime;
			  DECLARE $dateto      AS Datetime;
			  DECLARE $comment     AS String;
			  DECLARE $trans_cat   AS String;`
	switch repType {
	case HISTORY_BY_SUBCAT:
		query = query + `
select substring( cast(trans_date as String), 0, 7) as trans_cat --actually trans date
      ,trans_amount as t_sum
	  ,"" as comment
      ,trans_date
  from doc
 where trans_date between $datefrom and $dateto
   and trans_cat = $trans_cat
   and comment = $comment
   and client_id = $client_id
 order by trans_date DESC`
	case HISTORY_BY_CAT:
		query = query + `
select trans_cat --actually trans month
      ,SUM(trans_amount) as t_sum
	  ,"" as comment
  from (
    select SUBSTRING( CAST(trans_date as String), 0, 7) as trans_cat
          ,trans_amount
      from doc
     where trans_date between $datefrom and $dateto
       and trans_cat = $trans_cat
       and client_id = $client_id)
 group by trans_cat
 order by trans_cat DESC`
	case SUBCAT_TOTALS:
		query = query + `
select d.trans_cat as trans_cat
      ,"" as comment
      ,case d.direction
         when 0 then sum(d.trans_amount)
         else d.direction * sum(d.trans_amount) 
       end as t_sum
      ,sum(d.trans_amount) as ta, 1 as tp
  from doc d
 where d.trans_date between $datefrom and $dateto
   and d.trans_amount != 0
   and d.client_id = $client_id
 group by d.trans_cat, d.direction
 union all
select d.trans_cat as trans_cat
	  ,d.comment as comment
	  ,case d.direction
		 when 0 then sum(d.trans_amount)
		 else d.direction * sum(d.trans_amount) 
		 end as t_sum
	  ,sum(d.trans_amount) as ta, 2 as tp
  from doc d
 where d.trans_date between $datefrom and $dateto
   and d.trans_amount != 0
   and d.client_id = $client_id
 group by d.trans_cat, d.comment, d.direction
 order by trans_cat, tp ASC, ta DESC`
	case REGULAR_TOTALS:
		query = query + `
select d.trans_cat as trans_cat
		,"" as comment
		,case d.direction
			when 0 then sum(d.trans_amount)
			else d.direction * sum(d.trans_amount) 
		end as t_sum
		,sum(d.trans_amount) as ta, 1 as tp
	from doc d
	where d.trans_date between $datefrom and $dateto
	and d.trans_amount != 0
	and d.client_id = $client_id
	group by d.trans_cat, d.direction
union all
select case d.direction
			when -1 then 'Total debit'
			when 0  then 'Total deposit'
			else         'Total credit'
		end as trans_cat
		,"" as comment
		,case d.direction
			when 0 then sum(d.trans_amount)
			else d.direction * sum(d.trans_amount) 
		end as t_sum
		,sum(d.trans_amount) as ta, 2 as tp
	from doc d
	where d.trans_date between $datefrom and $dateto
	and d.trans_amount != 0
	and d.client_id = $client_id
	group by d.direction
	order by tp ASC, ta DESC;`
	}

	err = r.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			query,
			table.NewQueryParameters(
				table.ValueParam("$client_id", types.Uint64Value(p.User_id)),
				table.ValueParam("$datefrom", types.DatetimeValueFromTime(p.Date_from)),
				table.ValueParam("$dateto", types.DatetimeValueFromTime(p.Date_to)),
				table.ValueParam("$trans_cat", types.BytesValue([]byte(p.Add_attr1))),
				table.ValueParam("$comment", types.BytesValue([]byte(p.Add_attr2))),
			),
		)
		if err != nil {
			return
		}
		defer res.Close()
		if err = res.NextResultSetErr(ctx); err != nil {
			return
		}
		var (
			rn   sql.NullString
			rc   sql.NullString
			rv   sql.NullInt64
			name string
		)
		for res.NextRow() {
			err = res.ScanNamed(
				named.Optional("trans_cat", &rn),
				named.Optional("comment", &rc),
				named.Optional("t_sum", &rv),
			)
			if err != nil {
				return
			}
			switch repType {
			case SUBCAT_TOTALS:
				if rc.String == "" {
					name = rn.String + ":"
				} else {
					name = rc.String
				}
				rres = append(rres, ReportResult{
					Name: name,
					Val:  int(rv.Int64),
				})
			case REGULAR_TOTALS, HISTORY_BY_CAT, HISTORY_BY_SUBCAT:
				rres = append(rres, ReportResult{
					Name: rn.String,
					Val:  int(rv.Int64),
				})
			}
		}
		return res.Err() // for driver retry if not nil
	})
	return rres, err
}

type UserStats struct {
	ClientID      int   `json:"client_id"`
	AvgIncome     int64 `json:"income"`
	MonthWrkHours int64 `json:"month_work_hours"`
	AvgExpenses   int64 `json:"avg_expenses"`
	LowExpenses   int64 `json:"low_expenses"`
}

func (r *Ydb) GetUserStats(ctx context.Context, user_id int) (*UserStats, error) {
	stats := &UserStats{ClientID: user_id}
	err := r.Table().Do(ctx, func(ctx context.Context, s table.Session) (err error) {
		_, res, err := s.Execute(
			ctx,
			table.DefaultTxControl(),
			`DECLARE $client_id AS Uint64;
  SELECT JSON_VALUE(data, "$.avg_expenses"     RETURNING Int64) as avg_expenses,
         JSON_VALUE(data, "$.income"           RETURNING Int64) as income,
         JSON_VALUE(data, "$.low_expenses"     RETURNING Int64) as low_expenses,
         JSON_VALUE(data, "$.month_work_hours" RETURNING Int64) as month_work_hours
    FROM user_statistic
   WHERE client_id = $client_id;`,
			table.NewQueryParameters(table.ValueParam("$client_id", types.Uint64Value(uint64(user_id)))),
		)
		if err != nil {
			return err
		}
		defer res.Close()
		if err = res.NextResultSetErr(ctx); err != nil {
			return err
		}

		for res.NextRow() {
			err = res.ScanNamed(
				named.OptionalWithDefault("avg_expenses", &stats.AvgExpenses),
				named.OptionalWithDefault("income", &stats.AvgIncome),
				named.OptionalWithDefault("low_expenses", &stats.LowExpenses),
				named.OptionalWithDefault("month_work_hours", &stats.MonthWrkHours),
			)
			if err != nil {
				return err
			}
		}
		return res.Err() // for driver retry if not nil
	})

	if err != nil {
		return nil, err
	}

	return stats, nil
}
