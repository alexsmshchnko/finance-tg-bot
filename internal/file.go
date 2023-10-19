package internal

import (
	"fmt"
	"log"
	"strconv"
	"time"

	excelize "github.com/xuri/excelize/v2"
)

type ReceiptRec struct {
	time        time.Time
	category    string
	amount      int
	description string
}

const (
	configPage        = "conf"
	expensesCellStart = "A1"

	expensesFileName = "receipts.xlsx"
	expensesPage     = "Расходы"
)

func init() {
	f, err := OpenFile(expensesFileName)
	if err != nil {
		log.Fatalf("OpenFile err: %e", err)
	}
	defer f.Close()

	err = RefreshCategories(f)
	if err != nil {
		log.Fatalf("RefreshCategories err: %e", err)
	}

}

var (
	expenseCategories []string
	incomeCategories  []string
)

func GetExpenseCategories() (slc []string) {
	slc = expenseCategories
	fmt.Println(slc)
	return
}

func NewReceiptRec(time time.Time, category string, amount int, description string) (rec *ReceiptRec) {
	return &ReceiptRec{
		time:        time,
		category:    category,
		amount:      amount,
		description: description,
	}
}

func OpenFile(fileName string) (file *excelize.File, err error) {
	file, err = excelize.OpenFile(fileName)

	return
}

func getLastIdx(f *excelize.File) (idx string, err error) {
	last, err := f.GetRows(expensesPage)
	if err != nil {
		return
	}

	idx = strconv.Itoa(len(last) + 1)
	return
}

func AddNewExpense(rec *ReceiptRec) (err error) {
	f, err := OpenFile(expensesFileName)
	if err != nil {
		log.Fatalf("OpenFile err: %e", err)
	}
	defer f.Close()

	idx, err := getLastIdx(f)
	if err != nil {
		return
	}

	f.SetCellValue(expensesPage, "A"+idx, rec.time)
	f.SetCellValue(expensesPage, "B"+idx, rec.category)
	f.SetCellValue(expensesPage, "C"+idx, rec.amount)
	f.SetCellValue(expensesPage, "D"+idx, rec.description)

	err = f.Save()
	return
}

func RefreshCategories(f *excelize.File) (err error) {
	rows, err := f.Rows(configPage)
	if err != nil {
		return
	}

	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return err
		}
		for _, colCell := range row {
			expenseCategories = append(expenseCategories, colCell)
		}
	}

	return
}
