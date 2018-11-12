package main

import "sync"

// User structure
type User struct {
	id       string
	messages chan *Message
	commands chan *Message
	rooms    []string
	lock     *sync.RWMutex
}

// GetID return user id
func (u *User) GetID() string {
	return u.id
}

// GetMessages return a channel of *Message
func (u *User) GetMessages() <-chan *Message {
	return u.messages
}

// SendMessage sends a message to user
func (u *User) SendMessage(m *Message) *User {
	u.lock.RLock()
	defer u.lock.RUnlock()
	u.messages <- m
	return u
}

// RunCommand imitates user sending command to chat
func (u *User) RunCommand(m *Message) *User {
	u.lock.RLock()
	defer u.lock.RUnlock()
	u.commands <- m
	return u
}

// AttachRoom attaches room for list of these rooms, that user belongs to
func (u *User) AttachRoom(room string) {
	u.rooms = append(u.rooms, room)
}

// DetachRoom dettaches room from list of rooms, that user belongs to
func (u *User) DetachRoom(room string) {
	for i, v := range u.rooms {
		if v == room {
			u.rooms = append(u.rooms[:i], u.rooms[i+1:]...)
			break
		}
	}
}

// GetRooms return a list of rooms user belongs to
func (u *User) GetRooms() []string {
	return u.rooms
}
