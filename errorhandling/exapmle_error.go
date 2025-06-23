package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// Example 1: Basic Error Handling
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("Cannot divide by zero.")
	}
	return a / b, nil
}
``
// Example 2: Panic and Recover
func riskyOperation() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	fmt.Println("Starting risky operation.")
	panic("Something went wrong!")
	fmt.Println("This line will never be executed.")
}

// Example 3: Custom Errors
type CustomError struct {
	Function string
	Message  string
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Error in %s: %s", e.Function, e.Message)
}

func testFunction() error {
	return &CustomError{Function: "testFunction", Message: "Something bad happened"}
}

// Example 4: Logging Errors
func logError(err error) {
	file, fileErr := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if fileErr != nil {
		log.Fatalf("Failed to open error log file: %v", fileErr)
	}
	defer file.Close()

	// Print error to the file
	_, writeErr := fmt.Fprintf(file, "ERROR: %v\n", err)
	if writeErr != nil {
		fmt.Printf("Failed to write to error log file: %v\n", writeErr)
	}
}

func main() {
	// Example 1: Basic Error Handling
	fmt.Println("Example 1: Basic Error Handling")
	result, err := divide(4, 0)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Result:", result)
	}

	// Example 2: Panic and Recover
	fmt.Println("\nExample 2: Panic and Recover")
	riskyOperation()
	fmt.Println("Continuing execution after risky operation.")

	// Example 3: Custom Errors
	fmt.Println("\nExample 3: Custom Errors")
	err = testFunction()
	if err != nil {
		fmt.Println(err)
	}

	// Example 4: Logging Errors
	fmt.Println("\nExample 4: Logging Errors")
	err = errors.New("This is a test error")
	logError(err)
	fmt.Println("Open the 'files' at the top left and see the error logged to `error.log`.")
}
