package main

import (
	"fmt"
)

func main() {
	fmt.Println("this is my programming language")
	var five = 5
	var two = 2

	fmt.Println(five + two + 50)
	m := map[int]int{}

	fmt.Printf("the tenth fibonacci number is %d\n\n", fib(10, m))
	fmt.Printf("the sum of the first tenth fib numbers is %d \n\n", sumFib(10))

	fmt.Printf("%.5g", 4.6878959)

	fmt.Println("This is really unsafe")
}

func sumFib(n int) (sum int) {
	m := map[int]int{}

	for i := 1; i <= n; i++ {
		sum += fib(i, m)
	}

	return
}

func fib(n int, hmap map[int]int) int {

	if fibn, ok := hmap[n]; ok {
		return fibn
	}

	if n == 1 || n == 2 {
		return 1
	}

	fibn := fib(n-1, hmap) + fib(n-2, hmap)
	hmap[n] = fibn
	fmt.Printf("the %d fib number is %d\n", n, fibn)
	return fibn
}
