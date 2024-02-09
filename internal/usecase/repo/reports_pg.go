package repo

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type rep struct {
	category string
	amount   int
}

func (s *Repo) GetMonthReport(username, reptype string) (res string, err error) {
	var (
		query, amnt string
	)
	if reptype == "PREVMONTH" {
		query = `select case grouping(trans_cat)
						  when 1 then 'Total '
								    || case direction when 1 then 'credit' else 'debit' end
						  else trans_cat end           as trans_cat
					   ,direction * sum(trans_amount)  as t_sum
				   from document d
				  where d.trans_date between date_trunc('month', current_date - interval '1' month)
										 and date_trunc('month', current_date) - interval '1' day
					and d.client_id = $1
				  group by grouping sets ((trans_cat, direction), (direction)) 
				  order by grouping(trans_cat) asc, sum(trans_amount) desc`
	}

	data, err := s.Db.Query(query, username)
	if err != nil {
		return res, err
	}

	reprec := rep{}
	reprt := make([]rep, 0)

	for data.Next() {
		err = data.Scan(&reprec.category, &reprec.amount)
		if err != nil {
			return res, err
		}
		reprt = append(reprt, reprec)
	}

	printer := message.NewPrinter(language.Russian)
	str := strings.Builder{}
	for _, v := range reprt {
		amnt = printer.Sprintf("%d", v.amount)
		str.WriteString(fmt.Sprintf("|%-21s|%8s|\n", v.category, amnt))
	}

	return str.String(), err
}
