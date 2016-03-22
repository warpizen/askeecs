package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

type AEServer struct {
	db           *Database
	questions    *Collection
	users        *Collection
	tokens       map[string]*User
	ch_login     chan *Session
	ch_logout    chan string
	ch_getu      chan *AuthReq
	salts        map[string]string
	ch_newsalt   chan KVPair
	ch_delsalt   chan string
	ch_getsalt   chan StrResponse
	sessionStore *SessionStorer
	m            *mux.Router
}

type KVPair struct {
	Key, Val string
}

type StrResponse struct {
	Arg  string
	Resp chan string
}

type Session struct {
	Token string
	Who   *User
}

type AuthReq struct {
	Token string
	Ret   chan *User
}

func NewServer() *AEServer {
	s := new(AEServer)

	//Initialize database and collections
	s.db = NewDatabase("localhost:27017")
	s.questions = s.db.Collection("Questions", new(Question))
	s.users = s.db.Collection("Users", new(User))

	//Initiliaze others
	s.tokens = make(map[string]*User)
	s.salts = make(map[string]string)

	s.ch_delsalt = make(chan string)
	s.ch_getsalt = make(chan StrResponse)
	s.ch_getu = make(chan *AuthReq)
	s.ch_login = make(chan *Session)
	s.ch_logout = make(chan string)
	s.ch_newsalt = make(chan KVPair)

	s.m = mux.NewRouter()
	return s
}

func (s *AEServer) Init(secretfile string) {
	log.Println("AEServer Init called")
	//Setup cookie store for sessions
	// func main() { fmt.Println(base64.StdEncoding.EncodeToString(
	// securecookie.GenerateRandomKey(64))) }
	secret, err := ioutil.ReadFile(secretfile)
	if err != nil {
		panic(err)
	}
	//log.Println("read secret : " + string(secret))
	//decode := make([]byte, 128)
	decode, err1 := base64.StdEncoding.DecodeString(string(secret))
	if err1 != nil {
		panic(err1)
	}
	//log.Println("read decode secret : " + string(decode))
	s.sessionStore = NewSessionStorer("askeecs", decode)
	s.SetupRouting()
}

func (s *AEServer) Serve() {
	go s.SyncSessionRoutine()
	go s.SyncSaltRoutine()
	// server port
	if err := http.ListenAndServe(":1313", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func genRandString() string {
	buf := new(bytes.Buffer)
	io.CopyN(buf, rand.Reader, 32)
	return hex.EncodeToString(buf.Bytes())
}

func (s *AEServer) GetSessionToken() string {
	tok := genRandString()
	//Find a unique random string
	//for _,ok := s.tokens[tok]; ok; tok = genRandString() {}
	//Like, really? was i really checking to see if a 256 bit random
	//number was already in our cache? wow. (also, unsafe access of tokens)
	return tok
}

func (s *AEServer) FindUserByName(name string) *User {
	users := s.users.FindWhere(bson.M{"username": name})
	if len(users) == 0 {
		fmt.Println("User not found.")
		return nil
	}
	user, _ := users[0].(*User)
	return user
}

func (s *AEServer) GetAuthedUser(w http.ResponseWriter, r *http.Request) *User {
	//Verify user account or something
	s.sessionStore.SetParam(w, r)
	login, ok := s.sessionStore.Get("Login")
	if !ok {
		log.Printf("Invalid cookie!")
		return nil
	}

	user := s.syncGetUser(login.(string))
	if user == nil {
		log.Printf("Invalid cookie!")
		return nil
	}
	return user
}

func (s *AEServer) syncGetUser(token string) *User {
	a := new(AuthReq)
	a.Ret = make(chan *User)
	a.Token = token
	s.ch_getu <- a
	u := <-a.Ret
	return u
}

func (s *AEServer) SyncSessionRoutine() {
	for {
		select {
		case ses := <-s.ch_login:
			s.tokens[ses.Token] = ses.Who
		case log := <-s.ch_logout:
			if _, ok := s.tokens[log]; ok {
				delete(s.tokens, log)
			}
		case get := <-s.ch_getu:
			u, ok := s.tokens[get.Token]
			if !ok {
				get.Ret <- nil
			} else {
				get.Ret <- u
			}
		}
	}
}

func (s *AEServer) SyncSaltRoutine() {
	for {
		select {
		case gr := <-s.ch_getsalt:
			slt, ok := s.salts[gr.Arg]
			if !ok {
				gr.Resp <- ""
			} else {
				gr.Resp <- slt
			}
		case dsl := <-s.ch_delsalt:
			delete(s.salts, dsl)
		case add := <-s.ch_newsalt:
			s.salts[add.Key] = add.Val
		}
	}
}

func Message(s string) string {
	return Stringify(JM{"Message": s})
}

func DoHash(pass, salt string) string {
	h := sha256.New()
	h.Write([]byte(pass))
	h.Write([]byte(salt))
	return hex.EncodeToString(h.Sum(nil))
}
