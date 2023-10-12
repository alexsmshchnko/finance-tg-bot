package internal

import (
	"fmt"
	"strconv"
	"time"

	excelize "github.com/xuri/excelize/v2"
)

func OpenFile(file string) (err error) {
	const page = "Расходы"
	f, err := excelize.OpenFile(file)
	defer f.Close()
	fmt.Println(f.GetSheetMap())

	last, err := f.GetRows(page)
	fmt.Println(last)
	idx := strconv.Itoa(len(last) + 1)

	err = f.SetCellValue(page, "A"+idx, time.Now().Format("01/02/2006"))
	err = f.SetCellValue(page, "B"+idx, "развлечения")
	err = f.SetCellValue(page, "C"+idx, 100+len(last))
	//	fmt.Println(f)

	err = f.Save()

	return
}
