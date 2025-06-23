package main

import "fmt"


func main(){
	nums := []int{11,2,13,4,3}
	sum :=0

	for _ , nums := range nums{
		sum += nums;
	}

	fmt.Println("sum:", sum);

	for i , num := range nums {
		if num == 3{
			fmt.Println("index:", i)
		}
	}

	m := map[string]string{"a":"apple", "b": "ball"}

	for key , val := range m{
		fmt.Printf("%s -> %s\n", key, val)
	}

	for key := range m {
		fmt.Println("key:" ,key);
	}

	for i , c := range "azAZ"{
		fmt.Println(i,c);
	}
      
}