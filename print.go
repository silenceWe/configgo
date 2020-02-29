package configgo

import (
	"bytes"
	"fmt"
)

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
	var printBuf bytes.Buffer
	for k, v := range colMaxWidth {
		if k == 0 {
			printBuf.WriteString("┌")
		}
		printBuf.WriteString(printn("─", v))
		if k == len(colMaxWidth)-1 {
			printBuf.WriteString("┐")
			printLine(printBuf.String())
			printBuf.Reset()
		} else {
			printBuf.WriteString("┬")
		}
	}
	printBuf.Reset()
	// head2
	for k := range head {
		if k == 0 {
			printBuf.WriteString("│")
		}
		printBuf.WriteString(fillBlank(head[k], colMaxWidth[k]))
		printBuf.WriteString("│")
	}
	printLine(printBuf.String())
	printBuf.Reset()

	// head3
	for k, v := range colMaxWidth {
		if k == 0 {
			printBuf.WriteString("├")
		}
		printBuf.WriteString(printn("─", v))
		if k == len(colMaxWidth)-1 {
			printBuf.WriteString("┤")
			printLine(printBuf.String())
			printBuf.Reset()
		} else {
			printBuf.WriteString("┼")
		}
	}
	printBuf.Reset()

	// print rows
	rowNum := len(rows)
	for k, row := range rows {
		for colIndex, _ := range colMaxWidth {
			printBuf.WriteString("│")
			if len(row)-1 >= colIndex {
				printBuf.WriteString(fillBlank(row[colIndex], colMaxWidth[colIndex]))
			} else {
				printBuf.WriteString(printn(" ", colMaxWidth[colIndex]))
			}
		}
		printBuf.WriteString("│")
		printLine(printBuf.String())
		printBuf.Reset()

		if k == rowNum-1 {
			break
		}
		for k, v := range colMaxWidth {
			if k == 0 {
				printBuf.WriteString("├")
			}
			printBuf.WriteString(printn("─", v))
			if k >= len(colMaxWidth)-1 {
				printBuf.WriteString("┤")
				printLine(printBuf.String())
				printBuf.Reset()
			} else {
				printBuf.WriteString("┼")
			}
		}
	}

	printBuf.Reset()
	// tail
	for k, v := range colMaxWidth {
		if k == 0 {
			printBuf.WriteString("└")
		}
		printBuf.WriteString(printn("─", v))
		if k == len(colMaxWidth)-1 {
			printBuf.WriteString("┘")
			printLine(printBuf.String())
			printBuf.Reset()
		} else {
			printBuf.WriteString("┴")
		}
	}
	printBuf.Reset()
}

func printLine(s string) {
	// loggo.Infoln(s)
	fmt.Println(s)
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
