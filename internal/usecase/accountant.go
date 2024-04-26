package usecase

import (
	"context"
	"finance-tg-bot/internal/entity"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/hako/durafmt"
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

func (a *Accountant) GetCatsLimit(ctx context.Context, user_id int, limit string) (cats []entity.TransCatLimit, err error) {
	a.log.Debug("GetCatsLimit", "user_id", user_id, "limit", limit)
	cats, err = a.repo.GetCategories(ctx, user_id, limit)
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

	hourFloat := float64(60 * 60 * 1000 * 1000 * 1000)

	str := strings.Builder{}
	if (userStat.MonthWrkHours != 0 && userStat.AvgIncome != 0) || (userStat.AvgExpenses != 0 && userStat.LowExpenses != 0) {
		str.WriteString("Во времени:\n")
	}
	if userStat.MonthWrkHours != 0 && userStat.AvgIncome != 0 {
		wrkHours := time.Nanosecond * time.Duration(int(float64(userStat.MonthWrkHours)*hourFloat*float64(transAmount)/float64(userStat.AvgIncome)))
		wrkDays := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/float64(userStat.AvgIncome)))

		str.WriteString(fmt.Sprintf(" - рабочих часов: %s (%d часов в месяц)\n - зарабатывается за: %s (из среднего месячного расчета)\n",
			durafmt.Parse(wrkHours).LimitToUnit("hours").LimitFirstN(2).Format(units),
			userStat.MonthWrkHours,
			durafmt.Parse(wrkDays).LimitFirstN(2).Format(units)))
	}
	if userStat.AvgExpenses != 0 && userStat.LowExpenses != 0 {
		freeDays := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/float64(userStat.AvgExpenses)))
		economDays := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/float64(userStat.LowExpenses)))

		str.WriteString(fmt.Sprintf(" - можно прожить без работы: %s, без крупных трат: %s\n",
			durafmt.Parse(freeDays).LimitFirstN(2).Format(units),
			durafmt.Parse(economDays).LimitFirstN(2).Format(units)))
	}

	if userStat.AvgIncome > 0 {
		str.WriteString("\nНакопить:\n")
	}
	if userStat.AvgIncome > userStat.AvgExpenses && userStat.AvgExpenses > 0 {
		saveUp := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/(float64(userStat.AvgIncome-userStat.AvgExpenses))))

		str.WriteString(fmt.Sprintf(" - в обычном режиме за: %s\n",
			durafmt.Parse(saveUp).LimitFirstN(2).Format(units)))
	}
	if userStat.AvgIncome > userStat.LowExpenses && userStat.LowExpenses > 0 {
		saveUp := time.Nanosecond * time.Duration(int(30*24*hourFloat*float64(transAmount)/(float64(userStat.AvgIncome-userStat.LowExpenses))))

		str.WriteString(fmt.Sprintf(" - без крупных трат: %s\n",
			durafmt.Parse(saveUp).LimitFirstN(2).Format(units)))
	}

	str.WriteString(fmt.Sprintf(`
При инвестировании под 5%% (за вычетом инфляции): 
 - %d через 10 лет
 - %d через 25 лет
 - %d через 50 лет`,
		int(float64(transAmount)*math.Pow(float64(1.05), 10)),
		int(float64(transAmount)*math.Pow(float64(1.05), 25)),
		int(float64(transAmount)*math.Pow(float64(1.05), 50))))

	res = str.String()

	return
}
