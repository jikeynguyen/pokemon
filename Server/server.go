package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	HOST                 = "localhost"
	PORT                 = "8080"
	TYPE                 = "tcp"
	WORLD_X_SIZE         = 10
	WORLD_Y_SIZE         = 10
	MaxPokeonCount       = 75
	pokemonPerWave       = 15
	TIMER                = 30 * time.Second
	AliveTimerForPokemon = TIMER * 4
	EscapeINFINITELOOP   = 100
	MOVE_DISTANCE        = 1
	AUTO_CATCHING_TIMER  = 60 * time.Second
	POKEMON_PER_ACCOUNT  = 100
)

type Users struct {
	Users []User `json:"users"`
}

type Types struct {
	Types []Type `json:"types"`
}

type Pokemon struct {
	Species        string   `json:"species"`
	Typing         []string `json:"type"`
	Basestat       STAT     `json:"BaseStats"`
	Level          int      `json:"level"`
	AccumulatedEXP int      `json:"Exp"`
	timeSpawned    time.Time
}

type Pokemons struct {
	Pokemon []Pokemon `json:"pokemon"`
}

type Type struct {
	Typing   string   `json:"typing"`
	Resisted []string `json:"resisted"`
	Weakedto []string `json:"weakedto"`
	Neutral  []string `json:"neutral"`
	Immunity []string `json:"immue"`
}

type User struct {
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	OwnedPokemon []Pokemon `json:"pokemonsOwned"`
	XPOS         int
	YPOS         int
}

type STAT struct {
	BaseHP    int `json:"HP"`
	BaseATK   int `json:"ATK"`
	BaseDEF   int `json:"DEF"`
	BaseSPD   int `json:"SPD"`
	BaseSPATK int `json:"SATK"`
	BaseSPDEF int `json:"SDEF"`
}

var mu sync.Mutex
var exist_user Users
var exist_typing Types
var existing_pokemons Pokemons
var pokemonWiki = make(map[int]Pokemon)
var pokemonOnMap = make(map[[2]int]Pokemon) // new pokemon is mapped to an xy position

var activeUsers = make(map[int]User)
var userSelections = make(map[int][]Pokemon)

func main() {
	listener, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	defer listener.Close()

	fmt.Println("Server starting on port:" + PORT)

	userJson, err := os.Open("users.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer userJson.Close()
	userInfo, _ := io.ReadAll(userJson)
	json.Unmarshal(userInfo, &exist_user)

	typingJson, err := os.Open("pokedex.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer typingJson.Close()
	pokedexInfo, _ := io.ReadAll(typingJson)
	json.Unmarshal(pokedexInfo, &exist_typing)
	json.Unmarshal(pokedexInfo, &existing_pokemons)
	counter := 0
	for _, pkm := range existing_pokemons.Pokemon {
		pokemonWiki[counter] = pkm
		counter++
	}

	go func() {
		for {
			handleCreatePokemons()
		}
	}()

	go func() {
		for {
			time.Sleep(10 * time.Second)
			handleDeletePokemons()
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
doAction:
	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		request := string(buffer[:])
		fmt.Println(request)
		request_parts := strings.Split(request, " ")
		connectionID, _ := strconv.Atoi(request_parts[0])

		switch request_parts[1] {
		case "login":
			username_login := request_parts[2]
			pw_login := request_parts[3]
			validity := false
			var response string
			var newUser User
			for _, user := range exist_user.Users {
				if user.Username == username_login && user.Password == pw_login {
					validity = true
					newUser = user
					newUser.XPOS = -1
					newUser.YPOS = -1
					break
				}
			}
			if validity {
				for _, user := range activeUsers {
					if user.Username == username_login {
						validity = false
						response = "-1_User already logged in"
						break
					}
				}
			} else {
				response = "-1_Invalid username or password"
			}
			if validity {
				uID := rand.Intn(1000)
				_, isExist := activeUsers[uID]
				for isExist {
					uID = rand.Intn(1000)
					_, isExist = activeUsers[uID]
				}
				response = strconv.Itoa(uID) + "_Welcome user " + username_login
				activeUsers[uID] = newUser
			}
			conn.Write([]byte(response))

			if !validity {
				break doAction
			}

		case "logout":
			IDtoDelete, _ := strconv.Atoi(request_parts[2])
			delete(activeUsers, IDtoDelete)
			conn.Write([]byte("User has successfully logged out"))
			break doAction

		case "wiki":
			handleWiki(conn, request_parts)
		case "list":
			switch request_parts[2] {
			case "show":
				handleShowUserPokemons(conn, connectionID)
			}
		case "battle":
			handleBattle(conn, request_parts, connectionID)
		case "catch":
			handleCatch(conn, request_parts, connectionID)
		}
	}
	conn.Close()
}

func handleWiki(conn net.Conn, request_parts []string) {
	switch strings.ToLower(request_parts[2]) {
	case "mons":
		monToFind := strings.Title(strings.ToLower(request_parts[3]))
		var responseMss string
		pokemonIsExists := false
		for _, pkm := range existing_pokemons.Pokemon {
			if pkm.Species == monToFind {
				pokemonIsExists = true
				responseMss = "\nSpecies: " + pkm.Species + "\nType: "
				var whatType [2]string
				for index, pkmtype := range pkm.Typing {
					if pkmtype == "NONE" {
						whatType[index] = whatType[index-1]
					} else {
						whatType[index] = pkmtype
					}
				}
				if whatType[0] != whatType[1] {
					responseMss += whatType[0] + "/" + whatType[1] + "\nBasestat:"
				} else {
					responseMss += whatType[0] + "/" + whatType[1] + " (pure " + strings.ToLower(whatType[0]) + ")\nBasestat:"
				}

				responseMss += "\n-HP:" + strconv.Itoa(pkm.Basestat.BaseHP)
				responseMss += "\n-ATK:" + strconv.Itoa(pkm.Basestat.BaseATK)
				responseMss += "\n-DEF:" + strconv.Itoa(pkm.Basestat.BaseDEF)
				responseMss += "\n-SPD:" + strconv.Itoa(pkm.Basestat.BaseSPD)
				responseMss += "\n-SPECIAL ATK:" + strconv.Itoa(pkm.Basestat.BaseSPATK)
				responseMss += "\n-SPECIAL DEF:" + strconv.Itoa(pkm.Basestat.BaseSPDEF) + "\n"
				break
			}
		}

		if !pokemonIsExists {
			responseMss = "Cannot find this pokemon"
		}
		conn.Write([]byte(responseMss))
	case "types":
		typeToFind := strings.ToUpper(request_parts[3])
		var responseMss string
		typeIsExist := false
		for _, pkmtype := range exist_typing.Types {
			if pkmtype.Typing == typeToFind {
				typeIsExist = true
				responseMss = "Type: " + pkmtype.Typing

				responseMss += "\n\nImmunity to (x0): "
				if len(pkmtype.Immunity) > 0 {
					for _, imType := range pkmtype.Immunity {
						responseMss += "\n- " + imType
					}
				} else {
					responseMss += "NONE"
				}
				responseMss += "\n\nResisted to (x0.5): "

				if len(pkmtype.Resisted) > 0 {
					for _, nspType := range pkmtype.Resisted {
						responseMss += "\n- " + nspType
					}
				} else {
					responseMss += "NONE"
				}
				responseMss += "\n\nWeak against (x2): "
				if len(pkmtype.Weakedto) > 0 {
					for _, spType := range pkmtype.Weakedto {
						responseMss += "\n- " + spType
					}
				} else {
					responseMss += "NONE"
				}

				responseMss += "\n\nNeutral with (x1): "
				if len(pkmtype.Neutral) > 0 {
					for _, nType := range pkmtype.Neutral {
						responseMss += "\n- " + nType
					}
				} else {
					responseMss += "NONE"
				}
				break
			}
		}

		if !typeIsExist {
			responseMss = "Cannot find this type"
		}
		conn.Write([]byte(responseMss))

	default:
		conn.Write([]byte("Invalid search option"))
	}
}

func handleShowUserPokemons(conn net.Conn, connectionID int) {
	listString := "\nYour storage: "
	for index, pkm := range activeUsers[connectionID].OwnedPokemon {
		listString += "\n" + strconv.Itoa(index) + ". " + pkm.Species
		listString += "\tLv. " + strconv.Itoa(pkm.Level)
		listString += "\tExp: " + strconv.Itoa(pkm.AccumulatedEXP)
	}
	conn.Write([]byte(listString))
}

func handleCreatePokemons() {
	mu.Lock()
	for i := 0; i < pokemonPerWave; i++ {
		if len(pokemonOnMap) > MaxPokeonCount {
			break
		}

		//create new pkm to add to map
		whatToCreate := rand.Intn(len(existing_pokemons.Pokemon))
		var newPokemon Pokemon
		var pkmValidity = true
		loopCounter := 0
		newPokemon.Species = pokemonWiki[whatToCreate].Species
		newPokemon.Typing = pokemonWiki[whatToCreate].Typing
		newPokemon.Basestat = pokemonWiki[whatToCreate].Basestat
		newPokemon.Level = 1
		newPokemon.AccumulatedEXP = 0
		newPokemon.timeSpawned = time.Now()

		var position [2]int
		position[0] = rand.Intn(WORLD_X_SIZE)
		position[1] = rand.Intn(WORLD_Y_SIZE)
		positionValidity := true

		for usedPosition := range pokemonOnMap {
			if usedPosition[0] == position[0] && usedPosition[1] == position[1] {
				positionValidity = false
				break
			}
		}

		for !positionValidity {
			if loopCounter == EscapeINFINITELOOP {
				pkmValidity = false
				break
			}
			loopCounter++
			position[0] = rand.Intn(WORLD_X_SIZE)
			position[1] = rand.Intn(WORLD_Y_SIZE)

			for usedPosition := range pokemonOnMap {
				if usedPosition[0] == position[0] && usedPosition[1] == position[1] {
					positionValidity = false
					break
				}
			}
		}

		if pkmValidity {
			pokemonOnMap[position] = newPokemon
		} else {
			continue
		}
	}

	mu.Unlock()
	time.Sleep(TIMER)
}

func handleDeletePokemons() {
	currentTime := time.Now()
	for position, pkm := range pokemonOnMap {
		if currentTime.Sub(pkm.timeSpawned) >= AliveTimerForPokemon {
			mu.Lock()
			delete(pokemonOnMap, position)
			mu.Unlock()
		}
	}
}

func handleMovement(moveDir int, xpos *int, ypos *int) {
	switch moveDir {
	case 1: //left
		if *xpos-1 >= 0 {
			*xpos -= 1
		}
	case 2: // right
		if *xpos+1 < WORLD_X_SIZE {
			*xpos += 1
		}
	case 3: // up
		if *ypos-1 >= 0 {
			*ypos -= 1
		}
	case 4: // down
		if *ypos+1 < WORLD_Y_SIZE {
			*ypos += 1
		}
	}
}

func handleCatchMons(id int, uPosition [2]int) {
	if pkm, ok := pokemonOnMap[uPosition]; ok {
		user := activeUsers[id]
		if len(user.OwnedPokemon) < POKEMON_PER_ACCOUNT {
			var userIndex int
			for i, u := range exist_user.Users {
				if user.Username == u.Username {
					userIndex = i
					break
				}
			}
			pkmCatched := pokemonOnMap[uPosition]
			exist_user.Users[userIndex].OwnedPokemon = append(exist_user.Users[userIndex].OwnedPokemon, pkmCatched)
			delete(pokemonOnMap, uPosition)

			existingUserData, _ := json.MarshalIndent(exist_user, "", "\t")
			err := os.WriteFile("users.json", existingUserData, 0644)
			if err != nil {
				log.Fatal(err)
			}

			response := fmt.Sprintf("You have caught a %s\nType: %s\nBase Stats: HP=%d, ATK=%d, DEF=%d, SPD=%d, SATK=%d, SDEF=%d",
				pkm.Species, strings.Join(pkm.Typing, "/"), pkm.Basestat.BaseHP, pkm.Basestat.BaseATK, pkm.Basestat.BaseDEF, pkm.Basestat.BaseSPD, pkm.Basestat.BaseSPATK, pkm.Basestat.BaseSPDEF)
			conn, err := net.Dial(TYPE, HOST+":"+PORT)
			if err == nil {
				defer conn.Close()
				conn.Write([]byte(response))
			}
		}
	}
}

func handleBattle(conn net.Conn, request_parts []string, connectionID int) {
	opponentID, err := strconv.Atoi(request_parts[2])
	if err != nil {
		conn.Write([]byte("Invalid opponent ID"))
		return
	}

	user := activeUsers[connectionID]
	_, exists := activeUsers[opponentID]
	if !exists {
		conn.Write([]byte("Opponent not found"))
		return
	}

	userPokemonIndices := strings.Split(request_parts[3], ",")
	if len(userPokemonIndices) != 3 {
		conn.Write([]byte("You must select exactly 3 Pokémon"))
		return
	}

	var userPokemons []Pokemon
	for _, index := range userPokemonIndices {
		pokemonIndex, err := strconv.Atoi(index)
		if err != nil || pokemonIndex < 0 || pokemonIndex >= len(user.OwnedPokemon) {
			conn.Write([]byte("Invalid Pokemon index"))
			return
		}
		userPokemons = append(userPokemons, user.OwnedPokemon[pokemonIndex])
	}

	userSelections[connectionID] = userPokemons

	opponentPokemons, opponentSelected := userSelections[opponentID]
	if !opponentSelected {
		conn.Write([]byte("Waiting for opponent to select their Pokémon"))
		return
	}

	winner, analysis := simulateBattle(userPokemons, opponentPokemons)
	response := fmt.Sprintf("Battle result: %s vs %s. Winner: %s\n%s",
		formatTeam(userPokemons), formatTeam(opponentPokemons), winner, analysis)

	conn.Write([]byte(response))
	opponentConn, err := net.Dial(TYPE, HOST+":"+PORT)
	if err == nil {
		defer opponentConn.Close()
		opponentConn.Write([]byte(response))
	}

	delete(userSelections, connectionID)
	delete(userSelections, opponentID)
}

func formatTeam(team []Pokemon) string {
	var names []string
	for _, pkm := range team {
		names = append(names, pkm.Species)
	}
	return strings.Join(names, ", ")
}

func simulateBattle(team1, team2 []Pokemon) (string, string) {
	team1Power := 0
	team2Power := 0
	analysis := "Battle Analysis:\n"

	for i, pkm1 := range team1 {
		pkm2 := team2[i]
		power1 := pkm1.Basestat.BaseHP + pkm1.Basestat.BaseATK + pkm1.Basestat.BaseDEF + pkm1.Basestat.BaseSPD + pkm1.Basestat.BaseSPATK + pkm1.Basestat.BaseSPDEF
		power2 := pkm2.Basestat.BaseHP + pkm2.Basestat.BaseATK + pkm2.Basestat.BaseDEF + pkm2.Basestat.BaseSPD + pkm2.Basestat.BaseSPATK + pkm2.Basestat.BaseSPDEF

		team1Power += power1
		team2Power += power2

		analysis += fmt.Sprintf("Matchup %d: %s vs %s\n", i+1, pkm1.Species, pkm2.Species)
		analysis += fmt.Sprintf("Stats: %d vs %d\n", power1, power2)
		analysis += typeAnalysis(pkm1, pkm2) + "\n"
	}

	var winner string
	if team1Power > team2Power {
		winner = "Team 1"
	} else if team2Power > team1Power {
		winner = "Team 2"
	} else {
		winner = "Draw"
	}
	return winner, analysis
}

func typeAnalysis(pkm1, pkm2 Pokemon) string {
	var analysis string

	for _, t1 := range pkm1.Typing {
		for _, t2 := range pkm2.Typing {
			for _, typeInfo := range exist_typing.Types {
				if typeInfo.Typing == t1 {
					if contains(typeInfo.Weakedto, t2) {
						analysis += fmt.Sprintf("%s is weak to %s\n", pkm1.Species, pkm2.Species)
					} else if contains(typeInfo.Resisted, t2) {
						analysis += fmt.Sprintf("%s resists %s\n", pkm1.Species, pkm2.Species)
					} else if contains(typeInfo.Immunity, t2) {
						analysis += fmt.Sprintf("%s is immune to %s\n", pkm1.Species, pkm2.Species)
					}
				}
			}
		}
	}

	return analysis
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func handleCatch(conn net.Conn, request_parts []string, connectionID int) {
	var uPosition [2]int
	uID, _ := strconv.Atoi(request_parts[0])

	userX, _ := strconv.Atoi(request_parts[3])
	userY, _ := strconv.Atoi(request_parts[4])
	finished := false
	CatchValidity := true
	if userX == -1 || userY == -1 {
		userX = rand.Intn(WORLD_X_SIZE)
		userY = rand.Intn(WORLD_Y_SIZE)

		for _, user := range activeUsers {
			if user.XPOS == userX && user.YPOS == userY {
				CatchValidity = false
				break
			}
		}
		for !CatchValidity {
			userX = rand.Intn(WORLD_X_SIZE)
			userY = rand.Intn(WORLD_Y_SIZE)

			for _, user := range activeUsers {
				if user.XPOS == userX && user.YPOS == userY {
					CatchValidity = false
					break
				}
			}
		}
	}

	switch request_parts[2] {
	case "auto":
		go func() {
			time.Sleep(AUTO_CATCHING_TIMER)
			mu.Lock()
			finished = true
			mu.Unlock()
		}()

		for !finished {
			time.Sleep(time.Second)
			movementID := rand.Intn(4) + 1
			handleMovement(movementID, &userX, &userY)
			uPosition[0] = userX
			uPosition[1] = userY
			handleCatchMons(uID, uPosition)
		}
	case "manual":
		moveID, _ := strconv.Atoi(request_parts[5])
		handleMovement(moveID, &userX, &userY)
		uPosition[0] = userX
		uPosition[1] = userY
		handleCatchMons(uID, uPosition)
	}

	existingUserData, _ := json.MarshalIndent(exist_user, "", "\t")
	err := os.WriteFile("users.json", existingUserData, 0644)
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(request_parts[0] + " " + strconv.Itoa(userX) + " " + strconv.Itoa(userY) + " ."))
}
