package usecase

import (
	"context"
	"finance-tg-bot/internal/entity"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/hako/durafmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Accountant struct {
	repo     Repo
	user     User
	reporter Reporter
	sync     Cloud
	log      *slog.Logger
}

func New(d Repo, u User, r Reporter, s Cloud, l *slog.Logger) *Accountant {
	return &Accountant{
		repo:     d,
		user:     u,
		reporter: r,
		sync:     s,
		log:      l,
	}
}

func (a *Accountant) GetCatsLimit(ctx context.Context, user_id int) (cats []entity.TransCatLimit, err error) {
	a.log.Debug("GetCatsLimit", "user_id", user_id)
	cats, err = a.repo.GetCategories(ctx, user_id)
	if err != nil {
		a.log.Error("repo.GetCategories", "err", err)
	}
	return
}

func (a *Accountant) EditCats(ctx context.Context, tc *entity.TransCatLimit) (err error) {
	a.log.Debug("EditCats", "user_id", tc.UserId)
	err = a.repo.EditCategory(ctx, tc)
	if err != nil {
		a.log.Error("repo.EditCategory", "err", err)
	}
	return
}

func (a *Accountant) GetSubCats(ctx context.Context, user_id int, trans_cat string) (cats []string, err error) {
	a.log.Debug("GetSubCats", "user_id", user_id, "trans_cat", trans_cat)
	cats, err = a.repo.GetSubCategories(ctx, user_id, trans_cat)
	if err != nil {
		a.log.Error("repo.GetSubCategories", "err", err)
	}
	return
}

func (a *Accountant) GetUserStatus(ctx context.Context, username string) (id int, status bool, err error) {
	a.log.Debug("GetUserStatus request", "username", username)
	id, status, err = a.user.GetStatus(ctx, username)
	if err != nil {
		a.log.Error("user.GetStatus", "err", err)
	}
	a.log.Debug("GetUserStatus response", "username", username, "status", status)
	return
}

func (a *Accountant) PostDoc(ctx context.Context, doc *entity.Document) (err error) {
	a.log.Debug("PostDoc", "UserId", doc.UserId, "Category", doc.Category, "MsgID", doc.MsgID)
	err = a.repo.PostDocument(ctx, doc)
	if err != nil {
		a.log.Error("nrepo.PostDocument", "err", err)
	}
	return
}

func (a *Accountant) DeleteDoc(chat_id, msg_id string, user_id int) (err error) {
	a.log.Debug("DeleteDoc", "chat_id", chat_id, "msg_id", msg_id, "user_id", user_id)
	err = a.repo.DeleteDocument(context.Background(),
		&entity.Document{
			MsgID:  msg_id,
			ChatID: chat_id,
			UserId: user_id,
		},
	)
	if err != nil {
		a.log.Error("repo.DeleteDocument", "err", err)
	}
	return
}

func (a *Accountant) GetStatement(p map[string]string) (res string, err error) {
	a.log.Debug("GetStatementTotals", "p", p)
	res, err = a.reporter.GetStatementTotals(context.Background(), a.log, p)
	if err != nil {
		a.log.Error("reporter.GetStatementTotals", "err", err)
	}
	return
}

func (a *Accountant) Money2Time(transAmount int, user_id int) (res string, err error) {
	const (
		INVEST_PERCENT                  = 5
		INVEST_PERCENT_INCOME   float64 = 1 + float64(INVEST_PERCENT)/100
		DIVIDEND_REPCENT                = 4
		DIVIDEND_REPCENT_INCOME float64 = float64(DIVIDEND_REPCENT) / 100
	)
	transAmnt := float64(transAmount)

	a.log.Debug("GetUserStats", "user_id", user_id)
	userStat, err := a.reporter.GetUserStats(context.Background(), user_id)
	if err != nil {
		a.log.Error("reporter.GetUserStats", "err", err)
		return
	}

	units, err := durafmt.DefaultUnitsCoder.Decode("год:лет,неделя:недели,день:дней,час:часов,минута:минут,секунда:секунд,милисекунда:милисекунд,микросекунда:микросекунд")
	if err != nil {
		a.log.Error("durafmt.DefaultUnitsCoder.Decode", "err", err)
		return
	}

	p := message.NewPrinter(language.Russian)
	hourFloat := float64(60 * 60 * 1000 * 1000 * 1000)

	str := strings.Builder{}
	if (userStat.MonthWrkHours != 0 && userStat.AvgIncome != 0) || (userStat.AvgExpenses != 0 && userStat.LowExpenses != 0) {
		str.WriteString("🕰️ Во времени:\n")
	}
	if userStat.MonthWrkHours != 0 && userStat.AvgIncome != 0 {
		wrkHours := time.Nanosecond * time.Duration(int(float64(userStat.MonthWrkHours)*hourFloat*float64(transAmount)/float64(userStat.AvgIncome)))
		wrkDays := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/float64(userStat.AvgIncome)))

		str.WriteString(p.Sprintf(" • рабочих часов: %s (%d часов в месяц)\n • зарабатывается за: %s (из среднего месячного расчета)\n",
			durafmt.Parse(wrkHours).LimitToUnit("hours").LimitFirstN(2).Format(units),
			userStat.MonthWrkHours,
			durafmt.Parse(wrkDays).LimitFirstN(2).Format(units)))
	}
	if userStat.AvgExpenses != 0 && userStat.LowExpenses != 0 {
		freeDays := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/float64(userStat.AvgExpenses)))
		economDays := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/float64(userStat.LowExpenses)))

		str.WriteString(p.Sprintf(" • можно прожить без работы: %s, без крупных трат: %s\n",
			durafmt.Parse(freeDays).LimitFirstN(2).Format(units),
			durafmt.Parse(economDays).LimitFirstN(2).Format(units)))
	}

	if userStat.AvgIncome > 0 {
		str.WriteString("\n👛 Накопить:\n")
	}
	if userStat.AvgIncome > userStat.AvgExpenses && userStat.AvgExpenses > 0 {
		saveUp := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/(float64(userStat.AvgIncome-userStat.AvgExpenses))))

		str.WriteString(p.Sprintf(" • в обычном режиме за: %s\n",
			durafmt.Parse(saveUp).LimitFirstN(2).Format(units)))
	}
	if userStat.AvgIncome > userStat.LowExpenses && userStat.LowExpenses > 0 {
		saveUp := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/(float64(userStat.AvgIncome-userStat.LowExpenses))))

		str.WriteString(p.Sprintf(" • без крупных трат: %s\n",
			durafmt.Parse(saveUp).LimitFirstN(2).Format(units)))
	}

	if userStat.AvgExpenses != 0 && userStat.LowExpenses != 0 {
		str.WriteString(p.Sprintf(`
📈 При инвестировании под %d%% (за вычетом инфляции): 
 • %d (%s) через 10 лет
 • %d (%s) через 25 лет
 • %d (%s) через 50 лет`,
			INVEST_PERCENT,
			int(transAmnt*math.Pow(INVEST_PERCENT_INCOME, 10)),
			durafmt.Parse(time.Nanosecond*
				time.Duration(int(30*24*hourFloat*transAmnt/float64(userStat.AvgExpenses)*math.Pow(INVEST_PERCENT_INCOME, 10)))).
				LimitFirstN(2).Format(units),
			int(transAmnt*math.Pow(INVEST_PERCENT_INCOME, 25)),
			durafmt.Parse(time.Nanosecond*
				time.Duration(int(30*24*hourFloat*transAmnt/float64(userStat.AvgExpenses)*math.Pow(INVEST_PERCENT_INCOME, 25)))).
				LimitFirstN(2).Format(units),
			int(transAmnt*math.Pow(INVEST_PERCENT_INCOME, 50)),
			durafmt.Parse(time.Nanosecond*
				time.Duration(int(30*24*hourFloat*transAmnt/float64(userStat.AvgExpenses)*math.Pow(INVEST_PERCENT_INCOME, 50)))).
				LimitFirstN(2).Format(units)))
	} else {
		str.WriteString(p.Sprintf(`
📈 При инвестировании под %d%% (за вычетом инфляции): 
 • %d через 10 лет
 • %d через 25 лет
 • %d через 50 лет`,
			INVEST_PERCENT,
			int(transAmnt*math.Pow(INVEST_PERCENT_INCOME, 10)),
			int(transAmnt*math.Pow(INVEST_PERCENT_INCOME, 25)),
			int(transAmnt*math.Pow(INVEST_PERCENT_INCOME, 50))))
	}

	if userStat.AvgExpenses != 0 && userStat.LowExpenses != 0 {
		str.WriteString(p.Sprintf(`

💸 Дивидендная стратегия (%d%% за вычетом инфляции): 
 • %d (%s) каждый год
 • %d (%s) каждый месяц`,
			DIVIDEND_REPCENT,
			int(transAmnt*DIVIDEND_REPCENT_INCOME),
			durafmt.Parse(time.Nanosecond*
				time.Duration(int(30*24*hourFloat*transAmnt/float64(userStat.AvgExpenses)*DIVIDEND_REPCENT_INCOME))).
				LimitFirstN(2).Format(units),
			int(transAmnt*DIVIDEND_REPCENT_INCOME/12),
			durafmt.Parse(time.Nanosecond*
				time.Duration(int(30*24*hourFloat*transAmnt/float64(userStat.AvgExpenses)*DIVIDEND_REPCENT_INCOME/12))).
				LimitFirstN(2).Format(units),
		))
	} else {
		str.WriteString(p.Sprintf(`

💸 Дивидендная стратегия (%d%% за вычетом инфляции): 
 • %d каждый год
 • %d каждый месяц`,
			DIVIDEND_REPCENT,
			int(transAmnt*DIVIDEND_REPCENT_INCOME),
			int(transAmnt*DIVIDEND_REPCENT_INCOME/12),
		))
	}
	res = str.String()

	return
}
