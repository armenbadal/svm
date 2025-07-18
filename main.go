package main

import (
	"fmt"
	"os"
	"svm/assembler"
	"svm/machine"
)

func execute(input string) {
	_, err := os.Stat(input)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Ֆայլը գոյություն չունի կամ հասանելի չէ. %s\n", input)
		}
		return
	}

	bytes, err := assembler.Assemble(input)
	if err != nil {
		fmt.Printf("ՍԽԱԼ։ %s", err.Error())
		return
	}

	vm := machine.NewMachine()
	vm.Load(bytes)
	vm.Run()
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Ստեկային վիրտուալ մեքենա, v. 0.0.1")
		return
	}

	execute(os.Args[1])
}
