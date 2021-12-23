package main

import (
	"bufio"
	"fmt"
	"go-basic/basic"
	"os"
)

func main() {
	fmt.Print("Welcome to go-basic! Input command\n >")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() { // use `for scanner.Scan()` to keep reading
		input := scanner.Text()
		res, err := basic.Run(input, "stdin")
		if err != nil {
			fmt.Printf("Error! %s\n", err.Error())
		} else {
			fmt.Println(res.String())
		}
		fmt.Print(" >")
	}

}
