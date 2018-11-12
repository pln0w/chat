package main

import (
	"fmt"
	"time"

	"github.com/kr/pretty"
)

func runTest() {
	u1 := chat.CreateAndAttachUser()
	u2 := chat.CreateAndAttachUser()
	u3 := chat.CreateAndAttachUser()

	fmt.Printf("\n===> Added 3 users\n\n")

	pretty.Print(chat.users)

	chat.Subscribe(u1, "first")
	chat.Subscribe(u2, "first")
	chat.Subscribe(u3, "first")

	fmt.Printf("\n\n===> Subscribed them to one room\n\n")

	pretty.Print(chat.rooms)

	firstRoom := chat.rooms["first"]

	for _, user := range []*User{u1, u2, u3} {
		go func(u *User) {
			for {
				select {
				case msg := <-u.messages:

					if u.GetID() != msg.sender {
						tm := time.Unix(msg.createdAt, 0).Format(time.RFC3339)
						fmt.Printf("\n%s sees:\n\n[%s (%s): %s \n", u.GetID(), msg.sender, tm[11:][:8], msg.content)
					}
				}
			}
		}(user)
	}

	fmt.Printf("\n\n===> See how it goes when user %s sends message in room\n\n", u2.GetID())

	time.Sleep(time.Second * 2)

	// this should be done by a function broadcasting to room
	for _, ptr := range firstRoom.GetUsers() {
		ptr.SendMessage(&Message{
			sender:    u2.GetID(),
			content:   "I am u2 user and I just wanted to say hi",
			createdAt: time.Now().Unix(),
		})
	}

	time.Sleep(time.Second * 3)

	fmt.Printf("\n\n===> after while \n\n")

	// this should be done by a function broadcasting to room
	for _, ptr := range firstRoom.GetUsers() {
		ptr.SendMessage(&Message{
			sender:    u1.GetID(),
			content:   "Hi u2",
			createdAt: time.Now().Unix(),
		})
	}

	time.Sleep(time.Second * 2)
	fmt.Printf("\n\n===> after while \n\n")

	// this should be done by a function broadcasting to room
	for _, ptr := range firstRoom.GetUsers() {
		ptr.SendMessage(&Message{
			sender:    u3.GetID(),
			content:   "Nice to meet you!",
			createdAt: time.Now().Unix(),
		})
	}

	time.Sleep(time.Second * 1)

}
