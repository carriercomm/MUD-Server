package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

type ClientConnection struct {
	myConn    net.Conn
	myEncoder *gob.Encoder
	myDecoder *gob.Decoder

	//tradeChannel is used to easily communicate with the client without going
	//directly through the event manager. When used, it is expected that the event
	//manager has set the characters trading flag.
	tradeChannel chan string

	//pingChannel is used to easily receive pings from the client.
	pingChannel chan string

	character *Character

	//TODO this field can be removed since it is a global variable now
	EventManager *EventManager
}

//CliecntConnection constructor constructs a new client connection and sets the
//current event manager to the one supplied. This constructor is responsible
//for getting the initial room description.
func NewClientConnection(conn net.Conn, char *Character, decoder *gob.Decoder, encder *gob.Encoder) {
	cc := new(ClientConnection)
	cc.myConn = conn

	cc.myEncoder = encder
	cc.myDecoder = decoder

	cc.character = char
	char.myClientConn = cc
	cc.character.myClientConn = cc
	cc.EventManager = eventManager

	cc.tradeChannel = make(chan string)
	cc.pingChannel = make(chan string)

	go cc.Read()
}

//Read will continuously try to read from the connection established with the client.
//If an error occurs or the client wishes to exit this will handle closing the connection
//and cleaning up the character from the world. When a succesful read occures
//the corresponding event is added to the event queu or executed rightaway.
func (cc *ClientConnection) Read() {
	defer cc.shutdown()

	for {
		var clientResponse ClientMessage

		err := cc.myDecoder.Decode(&clientResponse)
		checkError(err, false)

		fmt.Println("Message read: ", clientResponse)

		if clientResponse.CombatAction {
			event := newEventFromMessage(clientResponse, cc.character)
			cc.EventManager.AddEvent(event)
		} else if clientResponse.getCommand() == "ping" {
			go cc.SendToPingChannel()
		} else if clientResponse.IsTradeCommand() {
			go cc.SendToTradeChannel(clientResponse)
		} else {
			go cc.EventManager.ExecuteNonCombatEvent(cc, &clientResponse)
		}

		if clientResponse.Command == "exit" || err != nil {
			fmt.Println("Closing the connection")
			break
		}
	}
}

func (cc *ClientConnection) shutdown() {
	cc.EventManager.RemovePlayerFromRoom(cc.getCharacter())
	cc.myConn.Close()
}

//Write will attempt to send the provided server message accross the connection
//to the client. Write automatically appends the clients most recent character
//data to the message.
func (cc *ClientConnection) Write(msg ServerMessage) {
	msg.addCharInfo(cc.character)
	err := cc.myEncoder.Encode(msg)
	checkError(err, false)
}

func (cc *ClientConnection) SendToTradeChannel(msg ClientMessage) {

	if msg.Command == "add" {
		for i := 0; i < msg.GetItemQuantity(); i++ {
			cc.tradeChannel <- msg.GetItem()
		}
	} else {
		cc.tradeChannel <- msg.GetValue()
	}
}

func (cc *ClientConnection) GetResponseToTrade() (response string) {
	timeoutChan := make(chan string)
	go func() {
		time.Sleep(time.Second * 60)
		timeoutChan <- "timeout"
	}()

	select {
	case msg := <-cc.tradeChannel:
		response = msg
	case msg := <-timeoutChan:
		response = msg
	}

	return response
}

func (cc *ClientConnection) GetItemsToTrade() string {
	return cc.GetResponseToTrade()
}

func (cc *ClientConnection) SendToPingChannel() {
	cc.pingChannel <- "ping"
}

func (cc *ClientConnection) GetResponseToPing(start time.Time) time.Duration {
	timeoutChan2 := make(chan string)
	go func() {
		time.Sleep(time.Second * 2)
		timeoutChan2 <- "timeout"
	}()

	select {
	case <-cc.pingChannel:
		return time.Now().Sub(start)
	case <-timeoutChan2:
		return time.Second * 6
	}
}

func (cc *ClientConnection) GetAverageRoundTripTime() (avg time.Duration) {

	for i := 0; i < 10; i++ {
		cc.Write(newServerMessageTypeS(PING, "ping"))
		avg += cc.GetResponseToPing(time.Now())
	}

	return ((avg / 10) / 2)
}

func (cc *ClientConnection) getCharactersName() string {
	return cc.character.Name
}

func (cc *ClientConnection) getCharactersRoomID() int {
	return cc.character.RoomIN
}

func (cc *ClientConnection) getCharacter() *Character {
	return cc.character
}

func (cc *ClientConnection) isConnectionClosed() bool {
	return cc.myConn == nil
}

func (cc *ClientConnection) giveItem(itm Item_I) {
	cc.character.AddItemToInventory(itm)
}
