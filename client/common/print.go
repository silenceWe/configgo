package common

import "fmt"

func TestTable() {
	head := []string{"col1", "col2", "col3"}
	rows := [][]string{
		[]string{"aa", "bb"},
		[]string{"aaadfa1", "bbadsfasd1"},
		[]string{"aa1", "badsfadsb1"},
		[]string{"asdfasdf", "badfasdfb1"},
		[]string{"aaadfa1", "bbadsfasd1"},
		[]string{"aa1", "badsfadsb1"},
		[]string{"asdfasdf", "badfasdfb1"},
		[]string{"aaadfa1", "bbadsfasd1"},
		[]string{"asdfasdf", "badfasdfb1"},
	}
	PrintTable(head, rows)
}

func PrintTable(head []string, rows [][]string) {
	maxColNum := 0
	colMaxWidth := make([]int, len(head))
	for k, v := range head {
		headLen := len(v)
		if headLen > colMaxWidth[k] {
			colMaxWidth[k] = headLen
		}
	}
	for _, row := range rows {
		colNum := len(row)
		if colNum > maxColNum {
			maxColNum = colNum
		}
		for colIndex, col := range row {
			colLen := len(col)
			if colLen > colMaxWidth[colIndex] {
				colMaxWidth[colIndex] = colLen
			}
		}
	}

	// head1
	for k, v := range colMaxWidth {
		if k == 0 {
			fmt.Print("┌")
		}
		fmt.Print(printn("─", v))
		if k == len(colMaxWidth)-1 {
			fmt.Print("┐\n")
		} else {
			fmt.Print("┬")
		}
	}
	// head2
	for k := range head {
		if k == 0 {
			fmt.Print("│")
		}
		fmt.Print(fillBlank(head[k], colMaxWidth[k]))
		fmt.Print("│")
	}
	fmt.Println()
	// head3
	for k, v := range colMaxWidth {
		if k == 0 {
			fmt.Print("├")
		}
		fmt.Print(printn("─", v))
		if k == len(colMaxWidth)-1 {
			fmt.Print("┤\n")
		} else {
			fmt.Print("┼")
		}
	}

	for rowIndex, row := range rows {
		for colIndex, _ := range colMaxWidth {
			fmt.Print("│")
			if len(row)-1 >= colIndex {
				fmt.Print(fillBlank(row[colIndex], colMaxWidth[colIndex]))
			} else {
				fmt.Print(printn(" ", colMaxWidth[colIndex]))
			}
		}
		fmt.Println("│")

		if (rowIndex+1)%3 == 0 && rowIndex != len(rows)-1 {
			for k, v := range colMaxWidth {
				if k == 0 {
					fmt.Print("├")
				}
				fmt.Print(printn("─", v))
				if k == len(colMaxWidth)-1 {
					fmt.Print("┤\n")
				} else {
					fmt.Print("┼")
				}
			}
		}
	}

	// tail
	for k, v := range colMaxWidth {
		if k == 0 {
			fmt.Print("└")
		}
		fmt.Print(printn("─", v))
		if k == len(colMaxWidth)-1 {
			fmt.Print("┘\n")
		} else {
			fmt.Print("┴")
		}
	}
}

func printn(s string, n int) string {
	r := ""
	for i := 0; i < n; i++ {
		r += s
	}
	return r
}
func fillBlank(s string, n int) string {
	l := len(s)
	if l >= n {
		return s
	}
	blank := printn(" ", n-l)
	return s + blank
}
