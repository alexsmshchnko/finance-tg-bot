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
	Time        time.Time
	Category    string
	Amount      int
	Description string
}

func (r ReceiptRec) String() string {
	return fmt.Sprintf("%v %s %d %s", r.Time, r.Category, r.Amount, r.Description)
}

const (
	configPage        = "conf"
	expensesCellStart = "A1"

	expensesPage = "Расходы"
)

func NewReceiptRec(time time.Time, category string, amount int, description string) (rec *ReceiptRec) {
	return &ReceiptRec{
		Time:        time,
		Category:    category,
		Amount:      amount,
		Description: description,
	}
}

func OpenFile(fileName string) (file *excelize.File, err error) {
	file, err = excelize.OpenFile(fileName)

	return
}

func GetRowsToSync(fileName string) (rslt []ReceiptRec, err error) {
	f, err := OpenFile(fileName)
	if err != nil {
		log.Fatalf("OpenFile err: %e", err)
	}
	defer f.Close()

	rows, err := f.GetRows("Расходы")

	var catR string
	var amntR int
	var desrR string
	var dt time.Time

	for i, v := range rows {
		if i < 16 {
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

		dt, _ = time.Parse("02.01.06", v[0])

		row := ReceiptRec{
			Time:        dt,
			Category:    catR,
			Amount:      amntR,
			Description: desrR,
		}

		rslt = append(rslt, row)
	}

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
