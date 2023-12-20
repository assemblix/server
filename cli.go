package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	welcomeMessage string = "Assemblix Server  Copyright (C) 2023 Assemblix.xyz"

	ps1 string = "> "
	ps2 string = ">> "
)

var cliBrodcast = make(chan string)

func cli() {
	fmt.Println(welcomeMessage)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(ps1)

		in, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}
		in = strings.TrimRight(in, "\n")

		var args = make([]string, 9)

		var index int
		var escape bool = false
		for _, c := range in {
			if index >= len(args) {
				break
			}
			switch c {
			case '\\':
				escape = true
			default:
				if escape {
					switch c {
					case 'n':
						args[index] += "\n"
					case '\\':
						args[index] += "\\"
					case ' ':
						args[index] += " "

					}
				} else if c == ' ' {
					index++
				} else {
					args[index] += string(c)
				}
			}
		}

		switch args[0] {
		case "":
		case "quit", "exit":
			return
		default:
			fmt.Printf("%s: command not found\n", args[0])
		}
	}
}
