package reports

import (
	"context"
	repPkg "finance-tg-bot/pkg/repository"
	"fmt"
	"log/slog"
	"strings"

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
	rows, err := r.repo.GetStatementTotals(ctx, log, p)
	if err != nil {
		return "", err
	}

	printer := message.NewPrinter(language.Russian)
	str := strings.Builder{}
	str.WriteString("+" + strings.Repeat("-", 22) + "+" + strings.Repeat("-", 8) + "+\n")
	for _, v := range rows {
		str.WriteString(fmt.Sprintf("|%-22s|%8s|\n", v.Name, printer.Sprintf("%d", v.Val)))
	}
	str.WriteString("+" + strings.Repeat("-", 22) + "+" + strings.Repeat("-", 8) + "+")

	return str.String(), err
}
