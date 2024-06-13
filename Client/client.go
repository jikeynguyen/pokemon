package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

var input = bufio.NewReader(os.Stdin)

var userID int
var xPos int = -1
var yPos int = -1

func main() {
	tcpServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	handleLogin(tcpServer)
}

func handleLogin(tcpServer *net.TCPAddr) {
	conn, err := net.DialTCP(TYPE, nil, tcpServer)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	fmt.Print("Welcome anonymous user!\nPlease enter your username and password to continue: ")
	uName_pass_login, err := input.ReadString('\n')
	if err != nil {
		fmt.Print(err)
		os.Exit(2)
	}
	parts := strings.Split(uName_pass_login, " ")
	if len(parts) != 2 {
		fmt.Println("Invalid input")

		_, err = conn.Write([]byte("login invalid invalid ."))
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		os.Exit(1)
	}

	parts[0] = strings.TrimSpace(parts[0])
	parts[1] = strings.TrimSpace(parts[1])

	hasher := sha256.New()
	hasher.Write([]byte(parts[1]))
	hash := hasher.Sum(nil)
	base64Hash := base64.URLEncoding.EncodeToString(hash)

	_, err = conn.Write([]byte("-1 login " + parts[0] + " " + base64Hash + " ."))
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	login_response := make([]byte, 1024)
	_, err = conn.Read(login_response)
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}

	user_valid := string(login_response)
	fmt.Println(user_valid)
	if strings.HasPrefix(user_valid, "-1") {
		os.Exit(0)
	} else {
		parts := strings.Split(user_valid, "_")
		IDstring := parts[0]
		userID, err = strconv.Atoi(IDstring)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		handleAction(conn)
	}
}

func handleAction(conn net.Conn) {
ACTION:
	for {
		fmt.Print("What do you want to do (wiki, list, battle, catch, logout): ")
		action, err := input.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}
		action = strings.TrimSpace(action)
		switch strings.ToLower(action) {
		case "wiki":
			fmt.Print("Search wiki with :mons/types + name\n-->")
			searchWhat, err := input.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				continue
			}
			searchWhat = strings.TrimSpace(searchWhat)
			_, err = conn.Write([]byte(strconv.Itoa(userID) + " wiki " + searchWhat + " ."))
			if err != nil {
				fmt.Println(err)
				continue
			}

			wiki_result := make([]byte, 10240)
			_, err = conn.Read(wiki_result)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(wiki_result))

		case "list":
			fmt.Print("Do you want to see your list of pokemons or delete one of them? (see / delete):")
			list_action, err := input.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				continue
			}

			list_action = strings.TrimSpace(list_action)
			switch list_action {
			case "see":
				_, err = conn.Write([]byte(strconv.Itoa(userID) + " list show ."))
				if err != nil {
					fmt.Println(err)
					continue
				}

				yourList := make([]byte, 10240)
				_, err = conn.Read(yourList)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(string(yourList))

			case "delete":
			}

		case "battle":
			fmt.Print("Enter opponent user ID: ")
			opponentID, err := input.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				continue
			}
			opponentID = strings.TrimSpace(opponentID)

			fmt.Println("Choose your 3 Pokémon for battle:")
			_, err = conn.Write([]byte(strconv.Itoa(userID) + " list show ."))
			if err != nil {
				fmt.Println(err)
				continue
			}

			yourList := make([]byte, 10240)
			_, err = conn.Read(yourList)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(yourList))

			fmt.Print("Enter the indices of 3 Pokémon to battle (comma-separated): ")
			userPokemonIndices, err := input.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				continue
			}
			userPokemonIndices = strings.TrimSpace(userPokemonIndices)
			indices := strings.Split(userPokemonIndices, ",")
			if len(indices) != 3 {
				fmt.Println("You must select exactly 3 Pokémon")
				continue
			}

			_, err = conn.Write([]byte(strconv.Itoa(userID) + " battle " + opponentID + " " + userPokemonIndices + " ."))
			if err != nil {
				fmt.Println(err)
				continue
			}

			battle_result := make([]byte, 1024)
			_, err = conn.Read(battle_result)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(battle_result))

		case "catch":
			if xPos != -1 && yPos != -1 {
				fmt.Printf("Your current cordination is: (%v, %v)\n", xPos, yPos)
			} else {
				fmt.Println("Welcome to PokeCatch, your position will be generated upon entering the world")
			}

			fmt.Print("Do you want to catch auto or manual? (auto / manual + direction): ")
			catch_How, err := input.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				continue
			}
			catch_How = strings.TrimSpace(catch_How)
			catch_How = strings.ToLower(catch_How)
			catch_How_parts := strings.Split(catch_How, " ")

			switch catch_How_parts[0] {
			case "auto":
				fmt.Println("do auto catch...")
				_, err := conn.Write([]byte(strconv.Itoa(userID) + " catch auto " + strconv.Itoa(xPos) + " " + strconv.Itoa(yPos) + " ."))
				if err != nil {
					fmt.Println(err)
					os.Exit(5)
				}

				server_response := make([]byte, 10240)
				_, err = conn.Read(server_response)
				if err != nil {
					fmt.Println(err)
					continue
				}
				catch_result := string(server_response[:])
				CRasParts := strings.Split(catch_result, " ")
				xPos, _ = strconv.Atoi(CRasParts[1])
				yPos, _ = strconv.Atoi(CRasParts[2])

				fmt.Println("Finish catching\n")

			case "manual":
				if len(catch_How_parts) != 2 {
					fmt.Println("Invalid catch format")
					continue
				}
				writeToServer := strconv.Itoa(userID) + " catch manual " + strconv.Itoa(xPos) + " " + strconv.Itoa(yPos)
				switch catch_How_parts[1] {
				case "left":
					writeToServer += " 1"
				case "right":
					writeToServer += " 2"
				case "up":
					writeToServer += " 3"
				case "down":
					writeToServer += " 4"
				default:
					fmt.Println("Invalid movement option")
					continue
				}
				_, err := conn.Write([]byte(writeToServer + " ."))
				if err != nil {
					fmt.Println(err)
					os.Exit(5)
				}

				server_response := make([]byte, 10240)
				_, err = conn.Read(server_response)
				if err != nil {
					fmt.Println(err)
					continue
				}
				catch_result := string(server_response[:])
				CRasParts := strings.Split(catch_result, " ")
				oldX := xPos
				oldY := yPos
				xPos, _ = strconv.Atoi(CRasParts[1])
				yPos, _ = strconv.Atoi(CRasParts[2])

				if oldX == xPos && oldY == yPos {
					fmt.Printf("Can't move %s, your position remains unchanged\nYour current position is: (%v, %v)\n\n", catch_How_parts[1], xPos, yPos)
				} else {
					fmt.Printf("You have successfully moved %s\nYour new position is: (%v, %v)\n\n", catch_How_parts[1], xPos, yPos)
				}

			default:
				fmt.Println("Invalid movement type")
			}

		case "logout":
			fmt.Println("Logging out...")

			_, err := conn.Write([]byte(strconv.Itoa(userID) + " logout " + strconv.Itoa(userID) + " ."))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			logout_response := make([]byte, 1024)
			_, err = conn.Read(logout_response)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(string(logout_response))
			break ACTION
		default:
			fmt.Println("Invalid action")
		}
	}
}
