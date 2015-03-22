package main

import (
	"encoding/xml"
	"github.com/daviddengcn/go-colortext"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
)

// this should be a stub that hold a connection to a client
// works like a thread on its own
type Character struct {
	Name         string
	RoomIN       int
	HitPoints    int
	Defense      int
	CurrentEM    *EventManager
	myClientConn *ClientConnection

	//	Strength int
	//	Constitution int
	//	Dexterity int
	//	Wisdom int
	//	Charisma int
	//	Inteligence int

	//	Location string

	//	Race string
	//	Class string

	PersonalInvetory Inventory

	//	Weapon Item
	ArmourSet map[string]Armour
}

func newCharacter(name string, room int, hp int, def int) *Character {
	char := new(Character)
	char.Name = name
	char.HitPoints = hp
	char.Defense = def
	char.PersonalInvetory = *newInventory()
	char.ArmourSet = make(map[string]Armour)

	return char
}
func newCharacterFromName(name string) *Character {

	loadCharacterData(name)

	return onlinePlayers[name]
}

func (c *Character) init(conn net.Conn, name string, em *EventManager) {

	c.Name = name
	c.setCurrentEventManager(em)
	c.myClientConn = new(ClientConnection)
	c.myClientConn.setConnection(conn)

}

func (c *Character) setCurrentEventManager(em *EventManager) {
	c.CurrentEM = em

}

func (c *Character) getEventMessage(msg ServerMessage) {
	//fmt.Print("I, ", (*c).Name, " receive msg : ")
	//fmt.Println(msg.Value)
	c.myClientConn.sendMsgToClient(msg)

}

func (c *Character) receiveMessage() {

	go c.routineReceiveMsg()
}

func (c *Character) routineReceiveMsg() {

	for {
		err := c.myClientConn.receiveMsgFromClient(c.CurrentEM)
		if err == io.EOF {
			//need to unsubscribe and let this character be devour by garbage collecter
			c.CurrentEM.unsubscribeListener(c)
			break
		}
	}
}

func (c *Character) getAttackRoll() int {
	return rand.Int() % 6
}

//TODO change some of these functions so that they return []FormatterString
// 		so the client can see the effects.

func (c *Character) wearArmor(location string, armr Armour) {
	if _, ok := c.ArmourSet[location]; ok { // already an item present
		//TODO
	} else {
		c.ArmourSet[location] = armr
		c.Defense += armr.defense
	}
}

func (c *Character) takeOffArmor(location string) {
	if _, ok := c.ArmourSet[location]; ok { // already an item present
		delete(c.ArmourSet, location)
	} else {
		//TODO
	}
}

func (c *Character) addItemToInventory(item Item) {
	c.PersonalInvetory.items[item.name] = item
}
func (char *Character) moveCharacter(direction string) []FormattedString {
	room := worldRoomsG[char.RoomIN]
	dirAsInt := convertDirectionToInt(direction)

	if room.Exits[dirAsInt] >= 0 {
		room.removePCFromRoom(char.Name)
		room.ExitLinksToRooms[dirAsInt].addPCToRoom(char.Name)
		char.RoomIN = room.Exits[dirAsInt]
		return room.ExitLinksToRooms[dirAsInt].getFormattedOutput()
	} else {
		output := make([]FormattedString, 1, 1)
		output[0].Color = ct.Black
		output[0].Value = "No exit in that direction"
		return output
	}
}

func (char *Character) makeAttack(targetName string) []FormattedString {
	target := worldRoomsG[char.RoomIN].getMonster(targetName)
	output := make([]FormattedString, 2, 2)

	a1 := char.getAttackRoll()
	if a1 >= target.Defense {
		target.HP -= 2
		output[0].Value = "\nYou hit the " + targetName + "!"
	} else {
		output[0].Value = "\nYou missed the " + targetName + "!"
	}

	a2 := target.getAttackRoll()
	if target.HP > 0 {
		if a2 >= char.Defense {
			char.HitPoints -= 1
			output[1].Value = "\nThe " + targetName + " hit you!"
		} else {
			output[1].Value = "\nThe " + targetName + " narrowly misses you!"
		}
	} else { //TODO add corpse to Rooms list of items
		// TODO  reward player exp
		output[1].Value = "\nThe " + targetName + " drops over dead."
		room := worldRoomsG[char.RoomIN]
		room.killOffMonster(targetName)
	}

	return output
}

type CharacterXML struct {
	XMLName xml.Name `xml:"Character"`
	Name    string   `xml:"Name"`
	RoomIN  int      `xml:"RoomIN"`
	HP      int      `xml:"HitPoints"`
	Defense int      `xml:"Defense"`
}

func loadCharacterData(charName string) {
	//TODO remove hard coding
	xmlFile, err := os.Open("C:\\Go\\src\\MUD-Server\\Characters\\" + charName + ".xml")
	checkError(err)
	defer xmlFile.Close()

	XMLdata, _ := ioutil.ReadAll(xmlFile)

	var charData CharacterXML
	xml.Unmarshal(XMLdata, &charData)

	char := newCharacter(charData.Name, charData.RoomIN, charData.HP, charData.Defense)
	onlinePlayers[charName] = char
}
