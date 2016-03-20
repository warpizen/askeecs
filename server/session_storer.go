package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kidstuff/mongostore"
	"gopkg.in/mgo.v2"
)

type SessionStorer struct {
	w          http.ResponseWriter
	r          *http.Request
	store      *mongostore.MongoStore
	db         *mgo.Session
	cookieName string
}

func NewSessionStorer(cookiename string, key []byte) *SessionStorer {
	log.Println("session cookie name : " + cookiename)
	o := &SessionStorer{cookieName: cookiename}
	// Fetch new store.
	var err error
	o.db, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	o.store = mongostore.NewMongoStore(o.db.DB(cookiename).C("session"),
		3600*24*365, true, key)
	return o
}

func (s SessionStorer) Close() {
	s.db.Close()
}

func (s *SessionStorer) SetParam(w http.ResponseWriter, r *http.Request) {
	s.w = w
	s.r = r
}

func (s SessionStorer) Get(key string) (interface{}, bool) {
	log.Println("SessionStorer Get : " + key)
	log.Println("SessionStorer cookiename : " + s.cookieName)
	session, err := s.store.Get(s.r, s.cookieName)
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	strInf, ok := session.Values[key]
	if !ok {
		return "", false
	}
	return strInf, true

	/*
		str, ok := strInf.(string)
		if !ok {
			return "", false
		}
	*/

	//return str, true
}

func (s SessionStorer) Put(key, value string) {
	log.Println("SessionStorer Put : " + key + " value : " + value)
	session, err := s.store.Get(s.r, s.cookieName)
	if err != nil {
		fmt.Println(err)
		return
	}

	session.Values[key] = value
	session.Save(s.r, s.w)
}

func (s SessionStorer) Del(key string) {
	log.Println("SessionStorer Del : " + key)
	session, err := s.store.Get(s.r, s.cookieName)
	if err != nil {
		fmt.Println(err)
		return
	}

	delete(session.Values, key)
	session.Save(s.r, s.w)
}
