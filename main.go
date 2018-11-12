package main

import (
	"C"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	charset        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	userIDLenght   = 5
	roomNameLength = 8
)

var (
	chat         *Chat
	randomSource *rand.Rand
	uiLog        chan string
	sentences    = map[int]string{
		0: "Hi! How are you?",
		1: "well, I'm not sure whether it's true",
		2: "Great!",
		3: "I'd like to deploy it asap :)",
		4: "you ain't gonna need it",
		5: "keep it simple, man",
		6: "let me check this",
		7: "is that a bug, or just a feature??",
		8: "guys! it, crashed!",
	}
)

func init() {
	randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// startUser adds a user to chat and generate it's behaviour
// by sending commands, creating conversations and sending messages
func startUser() {

	user := chat.CreateAndAttachUser()

	uiLog <- fmt.Sprintf("=> @%s joined chat", user.GetID())

	// serve communication handler for users
	go handleCommunication(user)

	for {

		// wait some rangom time for imitate normal behaviour
		time.Sleep((time.Duration(rand.Intn(4)) * 1000) * time.Millisecond)

		scenarios := map[int]int{
			0: 2,
			1: 4,
			2: 3,
			3: 1,
			4: 6,
		}

		var totalWeight int
		for _, g := range scenarios {
			totalWeight += g
		}

		randWeight := func(scenarios map[int]int, totalWeight int) int {
			rand.Seed(time.Now().UnixNano())
			r := rand.Intn(totalWeight)
			for _, g := range scenarios {
				r -= g
				if r <= 0 {
					return g
				}
			}
			return rand.Intn(len(scenarios) - 1)
		}

		switch randWeight(scenarios, totalWeight) {

		// 1st case is that user can see a list of users
		case 0:

			user.RunCommand(&Message{
				sender:    user.GetID(),
				content:   "/users",
				createdAt: time.Now().Unix(),
			})

		// 2nd case is that user can see a list of all conversations
		case 1:

			user.RunCommand(&Message{
				sender:    user.GetID(),
				content:   "/conversations",
				createdAt: time.Now().Unix(),
			})

		// 3rd case is that user can start conversations with some other users
		case 2:

			// ommit, if too many combinations
			if len(chat.rooms) > len(chat.users)*2 {
				continue
			}

			ru, err := chat.GetRandomUser()
			if nil != err {
				continue // not important here
			}

			rus, err := chat.GetRandomUsers(rand.Intn(len(chat.users) - 1))
			if nil != err {
				continue // not important here
			}

			userNames := ""
			for _, u := range rus {
				userNames += string(u.GetID()) + ", "
			}

			hasher := md5.New()
			hasher.Write([]byte(userNames))
			roomNameMD5 := hex.EncodeToString(hasher.Sum(nil))

			// should check here whether room already exist

			roomName := roomNameMD5[0:roomNameLength]

			go chat.Subscribe(ru, roomName)

			uiLog <- fmt.Sprintf("=> @%s joined %s", ru.GetID(), "#"+roomName)

			for _, u := range rus {
				if u.GetID() != ru.GetID() {
					go chat.Subscribe(u, roomName)
					uiLog <- fmt.Sprintf("=> @%s joined %s invited by @%s", u.GetID(), "#"+roomName, ru.GetID())
				}
			}

		// 4th case is that user can send message to others
		case 3:
			ru, err := chat.GetRandomUser()
			if nil != err {
				continue // not important here
			}

			rooms := ru.GetRooms()
			if len(rooms) > 1 {
				which := rand.Intn(len(rooms) - 1)
				var selectedRoom string
				for i, r := range rooms {
					if i == which {
						selectedRoom = r
					}
				}
				room := chat.rooms[selectedRoom]
				if room != nil && len(room.users) > 0 {
					others := room.users

					for _, ptr := range others {
						if ru.GetID() != ptr.GetID() {
							go func() {
								ptr.SendMessage(&Message{
									sender:    ru.GetID(),
									content:   sentences[rand.Intn(len(sentences)-1)],
									createdAt: time.Now().Unix(),
								})
							}()
						}
					}
				}

			}

		// 5th case is that user can leave conversation
		case 4:
			// ommit, if not enough rooms
			if len(chat.rooms) < (len(chat.users)/2)+1 {
				continue
			}

			ru, err := chat.GetRandomUser()
			if nil != err {
				continue // not important here
			}

			rooms := ru.GetRooms()
			if len(rooms) > 1 {
				which := rand.Intn(len(rooms) - 1)
				var selectedRoom string
				for i, r := range rooms {
					if i == which {
						selectedRoom = r
					}
				}
				chat.Unsubscribe(ru, selectedRoom)
				uiLog <- fmt.Sprintf("=> @%s leaved %s", ru.GetID(), "#"+selectedRoom)
			}
		}
	}
}

// handleCommunication is a function that is run per each user as goroutine
// to spawn two more goroutines concurrently listen for changes on its
// commands and messages channels and response
func handleCommunication(u *User) {

	// listen for commands
	go func() {
		for {
			select {
			case cmd := <-u.commands:

				tm := time.Unix(cmd.createdAt, 0).Format(time.RFC3339)
				uiLog <- fmt.Sprintf("[@%s (%s)]: %s", u.GetID(), tm[11:][:8], cmd.content)

				if cmd.content == "/users" {

					m := &Message{
						sender:    u.GetID(),
						content:   strings.Join(chat.GetUsersSlice(), ", "),
						createdAt: time.Now().Unix(),
					}
					u.SendMessage(m)
					uiLog <- fmt.Sprintf("[@bot => @%s]: %s \n", u.GetID(), m.content)

				}

				if cmd.content == "/conversations" {
					m := &Message{
						sender:    u.GetID(),
						content:   strings.Join(chat.GetRoomsSlice(), ", "),
						createdAt: time.Now().Unix(),
					}
					u.SendMessage(m)
					uiLog <- fmt.Sprintf("[@bot => @%s]: %s \n", u.GetID(), m.content)

				}
			}
		}
	}()

	// listen for messages
	go func() {
		for {
			select {
			case msg := <-u.messages:

				if u.GetID() != msg.sender {
					tm := time.Unix(msg.createdAt, 0).Format(time.RFC3339)
					uiLog <- fmt.Sprintf("[%s (%s): @%s %s \n", msg.sender, tm[11:][:8], u.GetID(), msg.content)

					// response sometimes
					if rand.Intn(10)%2 == 0 {
						go func() {
							if chat.users[msg.sender].GetID() != u.GetID() {
								chat.users[msg.sender].SendMessage(&Message{
									sender:    u.GetID(),
									content:   sentences[randomSource.Intn(len(sentences)-1)],
									createdAt: time.Now().Unix(),
								})
							}
						}()
					}
				}
			}
		}
	}()
}

func main() {
	numWorkers := flag.Int("users", 0, "number of users to generate")
	testMode := flag.Bool("manual", false, "run test")

	flag.Parse()

	// create chat object, that keeps all users and rooms
	chat = &Chat{
		users: make(map[string]*User),
		rooms: make(map[string]*Room),
		uLock: sync.RWMutex{},
		rLock: sync.RWMutex{},
	}

	// run only manual test an leave
	if *testMode == true {
		runTest()
		os.Exit(1)
	}

	if *numWorkers == 0 {
		fmt.Println("usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// allocate memory for string channel
	// this is used by ncsurses user interface to display all
	// events log text area.
	uiLog = make(chan string)

	for i := 0; i < *numWorkers; i++ {

		// add a user to chat and generate its behaviour
		go startUser()
	}

	// render console user interface
	stopRendering := make(chan os.Signal, 1)
	signal.Notify(stopRendering, os.Interrupt)

	go func() {
		for {
			select {
			case msg := <-uiLog:
				fmt.Printf("%s \n", msg)
			}
		}
	}()

	stopApp := make(chan os.Signal, 2)
	signal.Notify(stopApp, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stopApp
		os.Exit(0)
	}()
	<-stopApp
}
