// test
package main
import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"strings"
	"github.com/daviddengcn/go-colortext"
	_ "github.com/go-sql-driver/mysql"
)

func printFormatedOutput(output []FormattedString) {
	for _, element := range output {
		ct.ChangeColor(element.Color, false, ct.Black, false)
		fmt.Println(element.Value)
	}
	ct.ResetColor()
}

func MovementAndCombatTest(){
	var input string
	printFormatedOutput(worldRoomsG[0].getFormattedOutput())
	output := make([]FormattedString, 5, 5)
	for {		
		read, err := fmt.Scan(&input)
		checkError(err)	
		_ = read
		
		if(input == "exit"){
			break
		} else if (input == "attack") {
			var target string
			read, err = fmt.Scan(&target)
			checkError(err)	
			
			output = executeStandardAttack("Ragnar", target)
		} else { //assume movement
			output = onlinePlayers["Ragnar"].moveCharacter(input)
		}
	
		printFormatedOutput(output)
	}
}

func combatTest(){
	var input string
	var foo string
	printFormatedOutput(worldRoomsG[0].getFormattedOutput())
	read, err := fmt.Scanln(&input, &foo)
	checkError(err)	
		_ = read
	//fmt.Println(worldRoomsG[0].getFormattedOutput())	
	
	if( strings.HasPrefix(input, "attack") ) {
		//tmp := strings.Split(input, " ")
		//fmt.Println(tmp)
		output := executeStandardAttack("Ragnar", foo)
		printFormatedOutput(output)
	}
}

func roomAndMoveTest2(){
	var input string
	//fmt.Println(worldRoomsG[0].getFormattedOutput())
	printFormatedOutput(worldRoomsG[0].getFormattedOutput())
	for {
		//fmt.Println(rooms[currentRoom].Description, "\n")
		
		read, err := fmt.Scan(&input)
		checkError(err)	
		_ = read
		
		if(input == "exit"){
			break
		}
		
		output := onlinePlayers["Ragnar"].moveCharacter(input)
		//fmt.Println(output, "\n")
		printFormatedOutput(output)
	}
}

func roomAndMoveTest(rooms [4]*Room){
	currentRoom := 0
	var input string
	
	for {
		//fmt.Println(rooms[currentRoom].Description, "\n")
		fmt.Println(rooms[currentRoom].getFormattedOutput())
		read, err := fmt.Scan(&input)
		checkError(err)	
		_ = read
		
		if(input == "exit"){
			break
		}

		switch input {
			case "n", "N" : currentRoom = rooms[currentRoom].Exits[NORTH]
			case "s", "S" : currentRoom = rooms[currentRoom].Exits[SOUTH]
			case "w", "W" : currentRoom = rooms[currentRoom].Exits[WEST]
			case "e", "E" : currentRoom = rooms[currentRoom].Exits[EAST]
		}
	}
	
}

func LogInWithClientTest() {
	service := "127.0.0.1:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	conn, err := listener.Accept()
	checkError(err)
	fmt.Println("Connection established")
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	var clientsMessage ClientMessage
	decoder.Decode(&clientsMessage)

	fmt.Println("clients message: " + clientsMessage.Value)
	

	db, err := sql.Open("mysql",
		"admin1:admin@tcp(127.0.0.1:3306)/mud-database")
	checkError(err)
	defer db.Close()

	err = db.Ping()
	checkError(err)

	rows, err := db.Query("select CharacterNameLI, Password from Login where CharacterNameLI = ?", )
	defer rows.Close()
	var name string
	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(name)
	}

	checkError(rows.Err())

	reply := ServerMessage{Value: "This is the servers reply"}
	encoder.Encode(reply)
}

func LogInTest() {
	db, err := sql.Open("mysql",
		"admin1:admin@tcp(127.0.0.1:3306)/mud-database")
	checkError(err)
	defer db.Close()

	err = db.Ping()
	checkError(err)

	rows, err := db.Query("select CharacterNameLI from login")
	defer rows.Close()
	var name string
	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(name)
	}

	checkError(rows.Err())
}

func GobTest() {
	service := "127.0.0.1:1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	conn, err := listener.Accept()
	checkError(err)
	fmt.Println("Connection established")
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	var clientsMessage ClientMessage
	decoder.Decode(&clientsMessage)

	fmt.Println("clients message: " + clientsMessage.Value)
	reply := ServerMessage{Value: "This is the servers reply"}
	encoder.Encode(reply)

	conn.Close()
}

func databaseTest() {
	db, err := sql.Open("mysql",
		"admin1:admin@tcp(127.0.0.1:3306)/sakila")

	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	fmt.Println("After if stmt")
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("After 2nd if stmt")
	var name string
	rows, err := db.Query("select title from film")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(name)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func inventoryAndItemTest() {
	var i Item
	i.description = "a cool sword"
	i.name = "Short Sword"
	i.itemID = 666

	var i2 Item
	i2.description = "a cool stick"
	i2.name = "Oaken Bo"
	i2.itemID = 667

	items := [100]Item{i, i2}

	inventory := createInventory(items)

	var item = inventory.getItemByName("Short Sword")

	fmt.Println(item.name)
}
