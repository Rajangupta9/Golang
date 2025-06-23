package main

import "fmt"

// type Address struct{
// 	Name string
// 	City string
// 	Pincode int
// }

// type car struct{
// 	Name , Model , color string
// 	weightInKg       float64
// }

// func main(){
     
// 	var a Address
// 	a = Address{"rajan", "bihar", 8033118}
// 	fmt.Println(a)
    
// 	c := car{"ferrari", "GTC4" , "Red" , 1920}
// 	c.color = "black"
// 	fmt.Println(c);
// }

// Student struct with an anonymous structure and fields
type Student struct {
    personal struct {    // Anonymous inner structure for personal details
        name string
        enrollment int
    }
    GPA float64  // Standard field
}

func main() {
    student := Student{
        personal: struct {
            name string
            enrollment int
        }{
            name: "A",
            enrollment: 12345,
        },
        GPA: 3.8,
    }
    fmt.Println("Name:", student.personal.name)
    fmt.Println("Enrollment:", student.personal.enrollment)
    fmt.Println("GPA:", student.GPA)
}