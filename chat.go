package main

import (
	"errors"
	"math/rand"
	"reflect"
	"sort"
	"sync"
)

// Chat structure, messages broker
type Chat struct {
	users map[string]*User
	rooms map[string]*Room
	uLock sync.RWMutex
	rLock sync.RWMutex
}

// GetUsersSlice return a slice of users string representation
func (c *Chat) GetUsersSlice() []string {
	var users []string
	users = make([]string, 0)
	for _, u := range c.users {
		users = append(users, u.GetID())
	}
	sort.Strings(users)
	return users
}

// GetRoomsSlice return a slice of rooms string representation
func (c *Chat) GetRoomsSlice() []string {
	var rooms []string
	rooms = make([]string, 0)
	for _, r := range c.rooms {
		rooms = append(rooms, "#"+r.GetID())
	}
	sort.Strings(rooms)
	return rooms
}

// CreateAndAttachUser creates user object and adds him to the map
// with it's id as an unique key
func (c *Chat) CreateAndAttachUser() *User {
	c.uLock.Lock()
	defer c.uLock.Unlock()
	name := make([]byte, userIDLenght)
	for i := range name {
		name[i] = charset[randomSource.Intn(len(charset))]
	}
	u := &User{
		id:       string(name),
		messages: make(chan *Message),
		commands: make(chan *Message),
		rooms:    []string{},
		lock:     &sync.RWMutex{},
	}
	c.users[u.id] = u
	return u
}

// DetachAndRemoveUser unsubscribes user from all his rooms
// and removes its reference from chat users list
func (c *Chat) DetachAndRemoveUser(u *User) {
	c.uLock.Lock()
	defer c.uLock.Unlock()
	for _, r := range u.GetRooms() {
		c.Unsubscribe(u, r)
	}
	close(u.messages)
	delete(c.users, u.GetID())
}

// Subscribe adds user to the given room
func (c *Chat) Subscribe(u *User, room string) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	if nil == c.rooms[room] {
		c.rooms[room] = &Room{
			id: room,
		}
		c.rooms[room].users = make(map[string]*User)
	}
	c.rooms[room].users[u.GetID()] = u
	u.AttachRoom(room)
}

// Unsubscribe removes user from given room
func (c *Chat) Unsubscribe(u *User, room string) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	if nil != c.rooms[room] {
		if nil != c.rooms[room].users[u.GetID()] {
			delete(c.rooms[room].users, u.GetID())
			if 0 == len(c.rooms[room].users) {
				delete(c.rooms, room)
			}
			u.DetachRoom(room)
		}
	}
}

// GetRandomUsers return random slice of pointers to Users
func (c *Chat) GetRandomUsers(length int) ([]*User, error) {
	if length <= len(c.users) {
		var randomUsers []*User
		randomUsers = make([]*User, 0)
		for i := 0; i < length; i++ {
			keys := reflect.ValueOf(c.users).MapKeys()
			key := keys[rand.Intn(len(keys))].Interface()
			randomUsers = append(randomUsers, c.users[key.(string)])
		}
		return randomUsers, nil
	}
	return nil, errors.New("not enought users in chat")
}

// GetRandomUser return random pointer to User
func (c *Chat) GetRandomUser() (*User, error) {
	if len(c.users) > 1 {
		keys := reflect.ValueOf(c.users).MapKeys()
		key := keys[rand.Intn(len(keys))].Interface()

		return c.users[key.(string)], nil
	}
	return nil, errors.New("not enought users in chat")
}
