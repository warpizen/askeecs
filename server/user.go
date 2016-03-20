package main

import (
	"encoding/json"
	"io"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type User struct {
	ID       bson.ObjectId "_id,omitempty"
	Username string
	Password string `json:"-"`
	Salt     string `json:"-"`
	login    time.Time
}

func (u *User) New() I {
	return new(User)
}

func (u *User) GetID() bson.ObjectId {
	return u.ID
}

type AuthAttempt struct {
	Username string
	Password string
	Salt     string
}

func AuthFromJson(r io.Reader) *AuthAttempt {
	a := new(AuthAttempt)
	dec := json.NewDecoder(r)
	err := dec.Decode(a)
	if err != nil {
		return nil
	}
	return a
}

func (u *User) MakeComment(r io.Reader) *Comment {
	c := CommentFromJson(r)
	if c == nil {
		return nil
	}
	c.Author = u.Username
	c.ID = bson.NewObjectId()
	c.Timestamp = time.Now()
	return c
}

func (u *User) UpdateComment(r io.Reader, cid bson.ObjectId) *Comment {
	c := CommentFromJson(r)
	if c == nil {
		return nil
	}
	//log.Println(c)
	c.Author = u.Username
	c.ID = cid
	c.Timestamp = time.Now()
	return c
}

func (u *User) MakeResponse(r io.Reader) *Response {
	resp := ResponseFromJson(r)
	if r == nil {
		return nil
	}
	resp.Author = u.Username
	resp.ID = bson.NewObjectId()
	resp.Timestamp = time.Now()
	return resp
}

func (u *User) UpdateResponse(r io.Reader, rid bson.ObjectId) *Response {
	resp := ResponseFromJson(r)
	if r == nil {
		return nil
	}
	resp.Author = u.Username
	resp.ID = rid
	resp.Timestamp = time.Now()
	return resp
}
