package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	ps1 string = "> "
	ps2 string = ">> "
)

var cliBrodcast = make(chan string)

func cli() error {
	var license string = "Assemblix Server"
	func() {
		file, err := os.Open("LICENSE")
		if err != nil {
			logWarning(err)
			return
		}
		scanner := bufio.NewScanner(file)
		var line uint8 = 1
		for scanner.Scan() {
			if line == 3 {
				license += "  " + scanner.Text()
			}
			line++
		}
	}()
	fmt.Println(license)

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
		case "help":
			fmt.Println(" help       : show this menu")
			fmt.Println(" clear, cls : clear the screen")
			fmt.Println(" db         : open sqlite3")
			fmt.Println(" useradd    : create a new user ")
			fmt.Println(" quit, exit : exit the program")
		case "":
		case "quit", "exit":
			return nil
		case "clear", "cls":
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				fmt.Println(err)
				continue
			}
		case "db":
			cmd := exec.Command("sqlite3", "data.db")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Println(err)
				continue
			}
		case "useradd":
			var cash int = joinCash
			var admin bool = false
			var username string
			var password string

			var defaults int

			for i := 1; i < len(args); i++ {
				switch args[i] {
				case "-c", "--cash":
					if args[i+1] == "" {
						fmt.Printf("useradd: options %s: requires parameter\n", args[i])
						break
					}
					i++

					cashBuf, err := strconv.Atoi(args[i])
					if err != nil {
						fmt.Println(err)
						break
					}

					cash = cashBuf
				case "-a", "--admin":
					admin = true
				case "":
					break
				default:
					defaults++
					switch defaults {
					case 1:
						username = args[i]
					case 2:
						password = args[i]
					}
				}
			}

			if defaults < 2 {
				fmt.Println("useradd: not enough arguments")
				continue
			}
			if defaults > 2 {
				fmt.Println("useradd: too many arguments")
				continue
			}

			id, token, err := createUser(username, password, cash, admin, db)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Printf("id: %d\n", id)
			fmt.Printf("token: %s\n", token)
		default:
			fmt.Printf("%s: command not found\n", args[0])
		}
	}
}
