package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
	"webbugs-server/models"

	contracts "github.com/jental/webbugs-common-go/contracts"
	cmappers "github.com/jental/webbugs-common-go/mappers"
	cmodels "github.com/jental/webbugs-common-go/models"

	gosocketio "github.com/ambelovsky/gosf-socketio"
	transport "github.com/ambelovsky/gosf-socketio/transport"
	"github.com/google/uuid"
)

const port = ":5000"
const pageRadius = 30

var connectedClientCount uint = 0
var store Store = NewStore(pageRadius)

func onRegister(c *gosocketio.Channel) {
	newPlayerID := uuid.New()
	log.Printf("new player id: %v", newPlayerID)
	store.players[newPlayerID] = &models.PlayerInfo{
		ID:           newPlayerID,
		Name:         newPlayerID.String(),
		Client:       c,
		LastActivity: time.Now().UTC(),
	}

	allPlayerIDs := make([]uuid.UUID, len(store.players))
	i := 0
	for _, player := range store.players {
		allPlayerIDs[i] = player.ID
		i = i + 1
	}

	c.Emit(string(contracts.MetadataMessageType), contracts.MetadataContract{
		PlayerID:  newPlayerID,
		PlayerIDs: allPlayerIDs,
	})

	cc := make(chan cmodels.Coordinates)
	go store.field.GetRandomEmptyCellCoordinates(cc)
	crd := <-cc
	log.Printf("Setting a base at cell %v", crd)
	var event models.Event = models.NewSetBugEvent(
		crd,
		newPlayerID,
		true)
	store.Handle(&event)
}

func onClick(data contracts.ClickContract) {
	player, exists := store.players[data.PlayerID]
	if !exists {
		return
	}

	player.LastActivity = time.Now().UTC()

	event := models.NewClickEvent(
		cmodels.NewCoordinates(data.Crd.Cell.X, data.Crd.Cell.Y, data.Crd.Cell.Z),
		data.PlayerID)
	event2 := models.Event(event)
	store.Handle(&event2)
}

func onStoreUpdate() {
	for _, player := range store.players {
		if player.Client != nil {
			player.Client.Emit(string(contracts.DataMessageType), cmappers.MapFromData(store.field, store.components))
		}
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var savedGameToLoad string
	flag.StringVar(&savedGameToLoad, "l", "", "Specify a saved game file")
	flag.Parse()
	if savedGameToLoad != "" {
		log.Printf("Loading game: %v", savedGameToLoad)
		LoadSave(savedGameToLoad, &store)
	} else {
		log.Print("No game loaded. Epmty field will be used")
	}

	f, err := os.OpenFile("logs/out.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	store.subscribtions = append(store.subscribtions, onStoreUpdate)
	store.Start()
	// store.StartWallRemover()

	go store.startRemovingImmortals()

	server := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())

	server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Printf("socket client connected: %v", c.Ip())
		connectedClientCount++
	})

	server.On(gosocketio.OnError, func(c *gosocketio.Channel) {
		log.Printf("socket error: %v", c.Ip())
	})

	server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Printf("socket client closed: %v", c.Ip())
		connectedClientCount--
	})

	server.On(string(contracts.RegisterMessageType), func(c *gosocketio.Channel) {
		log.Printf("registration: %v", c.Ip())
		onRegister(c)
	})

	server.On(string(contracts.ClickMessageType), func(c *gosocketio.Channel, data contracts.ClickContract) {
		onClick(data)
	})

	server.On(string(contracts.ResetMessageType), func(c *gosocketio.Channel) {
		log.Printf("field reset: %v", c.Ip())
	})

	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir("../webbugs-client/dist/")))
	serveMux.Handle("/socket.io/", server)
	log.Printf("Serving at localhost%v...", port)
	log.Println(http.ListenAndServe(port, serveMux))
}
