package internal

import (
	"fmt"
	"log"

	excelize "github.com/xuri/excelize/v2"
)

const (
	configPage        = "conf"
	expensesCellStart = "A1"

	expensesFileName = "receipts.xlsx"
	expensesPage     = "Расходы"
)

var (
	expenseCategories []string
	incomeCategories  []string
)

func GetExpenseCategories() (slc []string) {
	slc = expenseCategories
	fmt.Println(slc)
	return
}

func init() {
	f, err := OpenFile(expensesFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = RefreshCategories(f)
	if err != nil {
		log.Fatal(err)
	}

}

func OpenFile(fileName string) (file *excelize.File, err error) {
	file, err = excelize.OpenFile(fileName)

	return

	// last, err := f.GetRows(expensesPage)
	// fmt.Println(last)
	// idx := strconv.Itoa(len(last) + 1)

	// err = f.SetCellValue(expensesPage, "A"+idx, time.Now().Format("01/02/2006"))
	// err = f.SetCellValue(expensesPage, "B"+idx, "развлечения")
	// err = f.SetCellValue(expensesPage, "C"+idx, 100+len(last))
	// //	fmt.Println(f)

	// err = f.Save()

	// return
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
			//fmt.Print(colCell, "\t")
		}
		//fmt.Println()
	}

	return
}
