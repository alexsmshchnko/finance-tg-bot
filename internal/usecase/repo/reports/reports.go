package reports

import (
	"context"
	"finance-tg-bot/internal/entity"
	repPkg "finance-tg-bot/pkg/repository"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Reports struct {
	repo repPkg.Reporter
}

func New(repPkg repPkg.Reporter) *Reports {
	return &Reports{repo: repPkg}
}

func (r *Reports) GetStatementTotals(ctx context.Context, log *slog.Logger, p map[string]string) (res string, err error) {
	var (
		rows []entity.ReportResult
	)
	rows, err = r.repo.GetStatementCatTotals(ctx, p)

	if err != nil {
		log.Error("repo.GetStatementTotals", "err", err)
		return
	}
	if len(rows) < 1 {
		log.Debug("got no rows from repo.GetStatementTotals")
		return "NO DATA", err
	}

	var nl, vl int
	printer := message.NewPrinter(language.Russian)
	for _, v := range rows {
		if nl < utf8.RuneCountInString(v.Name) {
			nl = utf8.RuneCountInString(v.Name)
		}
		if vl < utf8.RuneCountInString(printer.Sprintf("%d", v.Val)) {
			vl = utf8.RuneCountInString(printer.Sprintf("%d", v.Val))
		}
	}

	str := strings.Builder{}
	str.WriteString(strings.Repeat("-", nl+1) + "+" + strings.Repeat("-", vl+1) + "\n")
	for _, v := range rows {
		str.WriteString(fmt.Sprintf("%-"+strconv.Itoa(nl+1)+"s|%"+strconv.Itoa(vl+1)+"s\n", v.Name, printer.Sprintf("%d", v.Val)))
	}
	str.WriteString(strings.Repeat("-", nl+1) + "+" + strings.Repeat("-", vl+1) + "")

	return str.String(), err
}
