package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
)


func main(){
	csvFilename := flag.String("csv", "problems.csv", "a csv in the format of 'question, answer'")
	flag.Parse()
    file , err := os.Open(*csvFilename)
	

	if err != nil{
		exit(fmt.Sprintf("Failed to open the csv file %v", err))
	
	}
	r := csv.NewReader(file)

	lines, err := r.ReadAll()

	if err != nil{
		exit("failed to parsed the provided csv file.")
	}
    problem := parseLines(lines)
	

	correct :=0

	for i,p := range problem{
		fmt.Printf("Problem #%d: %s : ", i+1, p.q)
		var answer string
		fmt.Scanf("%s\n", &answer)
		if answer == p.a {
			fmt.Printf("Correct\n")
			correct++
		}else{
			fmt.Printf("Incorrect\n")
		}
	}
	fmt.Printf("your Score %d out of %d", correct , len(problem))
}

func parseLines(lines [][]string)[]problem{
	ret := make([]problem , len(lines))

	for i , line := range lines {
		ret[i] = problem{
			q: line[0],
			a: strings.TrimSpace(line[1]),
		}
	}
	return ret

}

type problem struct {
	q string
	a string
}
func  exit(msg string)  {
		fmt.Println(msg)
		os.Exit(1)
		
	}