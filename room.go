package main

import "fmt"

// Room structure
type Room struct {
	id    string
	users map[string]*User
}

// String is "toString()" Chat representation
func (r *Room) String() string {
	str := ""
	for _, u := range r.users {
		str += fmt.Sprintf("user: %s\n", (*u).GetID())
	}
	return str
}

// GetID return room id
func (r *Room) GetID() string {
	return r.id
}

// GetUsers return users
func (r *Room) GetUsers() map[string]*User {
	return r.users
}
