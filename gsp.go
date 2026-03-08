package main

import "fmt"

func generate(out chan int) {
	defer close(out)
	for i := 1; i <= 20; i++ {
		out <- i
	}
}

func square(in chan int, out chan int) {
	defer close(out)
	for got := range in {
		out <- got * got
	}
}

func drain(in chan int) {
	for got := range in {
		fmt.Println(got)
	}
}

func initGenerateSquarePrint() {
	generateOutCh := make(chan int)
	squareOutCh := make(chan int)

	go square(generateOutCh, squareOutCh)
	go generate(generateOutCh)
	drain(squareOutCh)
}
