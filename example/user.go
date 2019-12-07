package gen

import (
	"context"
	"github.com/Vliro/humus"
	"time"
)

//Save ensures password is saved only when it is a new user.
func (u *User) Save() humus.DNode {
	if u.Uid == "" {
		vals := u.Values().(*UserScalars)
		vals.Pass = u.Pass
		return vals
	}
	//Returning u.values() means we only want the scalar values.
	//return u.values()
	for k := range u.Friends {
		u.Friends[k].Reset()
		if u.Friends[k].FriendSince == nil {
			t := time.Now()
			u.Friends[k].FriendSince = &t
		}
	}
	return u
}

//GetUser returns user from database.
func GetUser(uid humus.UID) (*User, error) {
	var us User
	err := db.Query(context.Background(), humus.GetByUid(uid, UserFields), &us)
	return &us, err
}
