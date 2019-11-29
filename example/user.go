package gen

import (
	"context"
	"github.com/Vliro/mulbase"
)

//Save ensures password is saved only when it is a new user.
func (u *User) Save() mulbase.DNode {
	if u.Uid == "" {
		vals := u.Values().(*UserScalars)
		vals.Pass = u.Pass
		return vals
	}
	return u.Values()
}

//GetUser returns user from database.
func GetUser(uid mulbase.UID) (*User, error) {
	var us *User
	err := db.Query(context.Background(), mulbase.GetByUid(uid, UserFields), us)
	return us, err
}
