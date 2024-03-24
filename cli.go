package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/notwithering/argo"
)

const (
	ps1 string = "> "
)

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
				break
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

		args, invalid := argo.Parse(in)
		if invalid {
			continue
		}

		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "help":
			fmt.Println(" help       : show this menu")
			fmt.Println(" clear, cls : clear the screen")
			fmt.Println(" db         : open sqlite3")
			fmt.Println(" useradd    : create a new user : useradd [options...] username password")
			fmt.Println(" quit, exit : exit the program")
		case "":
		case "quit", "exit":
			return nil
		case "clear", "cls":
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				fmt.Printf("%s: %s", args[0], err)
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

		useradd:
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
					break useradd
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
