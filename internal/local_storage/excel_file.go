package local_storage

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	excelize "github.com/xuri/excelize/v2"
)

type excelFile struct {
	file *excelize.File
}

func New() *excelFile {
	return &excelFile{}
}

type ReceiptRec struct {
	Time        time.Time `json:"trans_date"`
	Category    string    `json:"trans_cat"`
	Amount      int       `json:"trans_amount"`
	Description string    `json:"comment"`
	Direction   int       `json:"direction"`
}

func (r ReceiptRec) String() string {
	return fmt.Sprintf("%v %s %d %s %d", r.Time, r.Category, r.Amount, r.Description, r.Direction)
}

const (
	configPage        = "conf"
	expensesCellStart = "A1"

	expensesPage = "Расходы"
)

func NewReceiptRec(time time.Time, category string, amount int, description string, direction int) (rec *ReceiptRec) {
	return &ReceiptRec{
		Time:        time,
		Category:    category,
		Amount:      amount,
		Description: description,
		Direction:   direction,
	}
}

func OpenFile(fileName string) (file *excelize.File, err error) {
	file, err = excelize.OpenFile(fileName)

	return
}

func getRows(f *excelize.File, page string, direction int) (rslt []ReceiptRec) {
	rows, err := f.GetRows(page)
	if err != nil {
		log.Fatalf("GetRows err: %e", err)
	}

	var catR string
	var amntR int
	var desrR string
	var dt time.Time

	for i, v := range rows {
		if i < 17 {
			continue
		}

		if len(v) > 1 {
			catR = v[1]
		}

		if len(v) > 2 {
			amntR, _ = strconv.Atoi(strings.TrimSpace(v[2]))
		}

		if len(v) > 3 {
			desrR = v[3]
		}

		dt, err = time.Parse("01-02-06", v[0])
		if err != nil {
			dt, err = time.Parse("02/01/2006", v[0])
			if err != nil {
				dt, _ = time.Parse("02.01.06", v[0])
			}
		}

		if dt.After(time.Now()) {
			continue
		}

		row := ReceiptRec{
			Time:        dt,
			Category:    catR,
			Amount:      amntR,
			Description: desrR,
			Direction:   direction,
		}

		rslt = append(rslt, row)
	}
	return
}

func GetRowsToSync(fileName string) (rslt []ReceiptRec, err error) {
	f, err := OpenFile(fileName)
	if err != nil {
		log.Fatalf("OpenFile err: %e", err)
	}
	defer f.Close()

	rslt = append(rslt, getRows(f, "Расходы", -1)...)
	rslt = append(rslt, getRows(f, "Доходы", 1)...)

	return
}

func getLastIdx(f *excelize.File) (idx int, err error) {
	last, err := f.GetRows(expensesPage)
	if err != nil {
		return
	}

	//idx = strconv.Itoa(len(last) + 1)
	return len(last), err
}

func AddNewExpense(rec *ReceiptRec, fileName string) (err error) {
	f, err := OpenFile(fileName)
	if err != nil {
		log.Fatalf("OpenFile err: %e", err)
	}
	defer f.Close()

	idx, err := getLastIdx(f)
	if err != nil {
		return
	}
	sIdx := strconv.Itoa(idx + 1)

	dateStr := rec.Time.Format("2006-01-02")
	dateDt, _ := time.Parse("2006-01-02", dateStr)

	f.SetCellValue(expensesPage, "A"+sIdx, dateDt)
	f.SetCellValue(expensesPage, "B"+sIdx, rec.Category)
	f.SetCellValue(expensesPage, "C"+sIdx, rec.Amount)
	//f.SetCellValue(expensesPage, "D"+idx, rec.description)1

	err = f.Save()
	return
}

func DeleteLastExpense(fileName string) (err error) {
	f, err := OpenFile(fileName)
	if err != nil {
		log.Fatalf("OpenFile err: %e", err)
	}
	defer f.Close()

	idx, err := getLastIdx(f)
	if err != nil {
		return
	}
	err = f.RemoveRow(expensesPage, idx)
	if err != nil {
		return
	}

	err = f.Save()
	return
}

func EditLastExpense(rec *ReceiptRec, fileName string) (err error) {
	f, err := OpenFile(fileName)
	if err != nil {
		log.Fatalf("OpenFile err: %e", err)
	}
	defer f.Close()

	idx, err := getLastIdx(f)
	if err != nil {
		return
	}
	sIdx := strconv.Itoa(idx)
	f.SetCellValue(expensesPage, "A"+sIdx, rec.Time)
	f.SetCellValue(expensesPage, "B"+sIdx, rec.Category)
	f.SetCellValue(expensesPage, "C"+sIdx, rec.Amount)
	f.SetCellValue(expensesPage, "D"+sIdx, rec.Description)

	err = f.Save()
	return

}

func AddLastExpenseDescription(rec *ReceiptRec, fileName string) (err error) {
	f, err := OpenFile(fileName)
	if err != nil {
		log.Fatalf("OpenFile err: %e", err)
	}
	defer f.Close()

	idx, err := getLastIdx(f)
	if err != nil {
		return
	}
	sIdx := strconv.Itoa(idx)
	f.SetCellValue(expensesPage, "D"+sIdx, rec.Description)

	err = f.Save()
	return

}
