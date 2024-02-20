package repo

import (
	"finance-tg-bot/internal/entity"
	"fmt"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type rep struct {
	category string
	amount   int
}

func (s *Repo) GetStatement(p *entity.Report) (res string, err error) {
	var query, amnt string
	switch p.RepName {
	case "TotalsForThePeriod":
		query = `
select case grouping(trans_cat)
		 when 1 then 'Total '
		     	   || case direction when 1 then 'credit' else 'debit' end
		 else trans_cat end           as trans_cat
	  ,direction * sum(trans_amount)  as t_sum
  from document d
 where d.trans_date between to_date($1,'dd.mm.yyyy') and to_date($2,'dd.mm.yyyy')
   and d.client_id = $3
 group by grouping sets ((trans_cat, direction), (direction)) 
 order by grouping(trans_cat) asc, sum(trans_amount) desc`
	}

	data, err := s.Query(
		query,
		p.RepParms["datefrom"],
		p.RepParms["dateto"],
		p.RepParms["username"],
	)
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
	str.WriteString("+" + strings.Repeat("-", 22) + "+" + strings.Repeat("-", 8) + "+\n")
	for _, v := range reprt {
		amnt = printer.Sprintf("%d", v.amount)
		str.WriteString(fmt.Sprintf("|%-22s|%8s|\n", v.category, amnt))
	}
	str.WriteString("+" + strings.Repeat("-", 22) + "+" + strings.Repeat("-", 8) + "+")

	return str.String(), err
}
