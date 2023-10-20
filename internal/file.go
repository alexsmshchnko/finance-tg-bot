package internal

import (
	"fmt"
	"log"
	"strconv"
	"time"

	excelize "github.com/xuri/excelize/v2"
)

type FileConfig struct {
	fileName string

	expensesPage      string
	expensesConfPage  string
	expensesCellStart string
	expensesCellEnd   string

	incomePage      string
	incomeConfPage  string
	incomeCellStart string
	incomeCellEnd   string
}

func NewFileConfig(fileName string) *FileConfig {
	return &FileConfig{
		fileName: fileName,
	}
}

type ReceiptRec struct {
	time        time.Time
	category    string
	amount      int
	Description string
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

func GetIncomeCategpries() (slc []string) {
	slc = incomeCategories
	fmt.Println(slc)
	return
}

func NewReceiptRec(time time.Time, category string, amount int, description string) (rec *ReceiptRec) {
	return &ReceiptRec{
		time:        time,
		category:    category,
		amount:      amount,
		Description: description,
	}
}

func OpenFile(fileName string) (file *excelize.File, err error) {
	file, err = excelize.OpenFile(fileName)

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
	sIdx := strconv.Itoa(idx + 1)

	f.SetCellValue(expensesPage, "A"+sIdx, rec.time)
	f.SetCellValue(expensesPage, "B"+sIdx, rec.category)
	f.SetCellValue(expensesPage, "C"+sIdx, rec.amount)
	//f.SetCellValue(expensesPage, "D"+idx, rec.description)

	err = f.Save()
	return
}

func EditLastExpense(rec *ReceiptRec) (err error) {
	f, err := OpenFile(expensesFileName)
	if err != nil {
		log.Fatalf("OpenFile err: %e", err)
	}
	defer f.Close()

	idx, err := getLastIdx(f)
	if err != nil {
		return
	}
	sIdx := strconv.Itoa(idx)
	f.SetCellValue(expensesPage, "A"+sIdx, rec.time)
	f.SetCellValue(expensesPage, "B"+sIdx, rec.category)
	f.SetCellValue(expensesPage, "C"+sIdx, rec.amount)
	f.SetCellValue(expensesPage, "D"+sIdx, rec.Description)

	err = f.Save()
	return

}

func AddLastExpenseDescription(rec *ReceiptRec) (err error) {
	f, err := OpenFile(expensesFileName)
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
		expenseCategories = append(expenseCategories, row...)
	}

	return
}
