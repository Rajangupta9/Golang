package view

import (
	"fmt"
	"strconv"
	"strings"
	"task-cli/datamodel"
)

// PromptTableTasks prints in the console a table
// with a header and rows containing all information
// of the tasks provided in the function
func PromptTableTasks(tasks []datamodel.Task) {
	headers := []string{"ID", "Status", "Description"}
	var rows [][]string

	for _, task := range tasks {
		rows = append(rows, []string{
			strconv.Itoa(task.ID),
			task.Status,
			task.Description,
		})
	}

	printTable(headers, rows)
}

// printTable creates a table with header and rows
func printTable(headers []string, rows [][]string) {
	// Calculate the width of each column
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, col := range row {
			if len(col) > colWidths[i] {
				colWidths[i] = len(col)
			}
		}
	}

	// Print the header row
	printRow(headers, colWidths)
	printSeparator(colWidths)

	// Print each data row
	for _, row := range rows {
		printRow(row, colWidths)
	}
}

// printRow prints a row with proper alignment
func printRow(row []string, colWidths []int) {
	for i, col := range row {
		fmt.Printf("| %-*s ", colWidths[i], col)
	}
	fmt.Println("|")
}

// printSeparator prints a separator line
func printSeparator(colWidths []int) {
	for _, width := range colWidths {
		fmt.Printf("+-%s-", strings.Repeat("-", width))
	}
	fmt.Println("+")
}
