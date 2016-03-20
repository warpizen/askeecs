package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

type handleFunc func(http.ResponseWriter, *http.Request) (int, string)

type wrapperHandler struct {
	s *AEServer
	f handleFunc
}

func wrapperHandle(s *AEServer, f handleFunc) wrapperHandler {
	return wrapperHandler{s, f}
}

func (wg wrapperHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	/*
		code, value := wg.f(w, r)
		w.WriteHeader(code)
		io.WriteString(w, value)
	*/
	if wg.s == nil {
		code, value := wg.f(w, r)
		w.WriteHeader(code)
		io.WriteString(w, value)
	} else {
		user := wg.s.GetAuthedUser(w, r)
		if user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, Message("Not authorized to .."))
		} else {
			code, value := wg.f(w, r)
			w.WriteHeader(code)
			io.WriteString(w, value)
		}
	}
}

func (s *AEServer) SetupRouting() {
	log.Println("SetupRouteing() - called")
	gets := s.m.Methods("GET").Subrouter()
	puts := s.m.Methods("PUT").Subrouter()
	posts := s.m.Methods("POST").Subrouter()
	//gets.Handle("/test", wrapperHandle(s, s.HandleNotFound))
	posts.Handle("/register", wrapperHandle(nil, s.HandleRegister))
	posts.Handle("/login", wrapperHandle(nil, s.HandleLogin))
	posts.Handle("/salt", wrapperHandle(nil, s.HandleGetSalt))
	posts.Handle("/register/salt", wrapperHandle(nil, s.HandleUniqueSalt))
	posts.Handle("/logout", wrapperHandle(nil, s.HandleLogout))
	posts.Handle("/me", wrapperHandle(s, s.HandleMe))
	gets.Handle("/maxpage", wrapperHandle(s, s.HandleGetDBCount))
	// get question list
	gets.Handle("/q", wrapperHandle(s, s.HandleGetQuestions))
	// insert question
	posts.Handle("/q", wrapperHandle(s, s.HandlePostQuestion))
	// get question
	gets.Handle("/q/{id}", wrapperHandle(s, s.HandleGetQuestion))
	// update question
	posts.Handle("/q/{id}", wrapperHandle(s, s.HandleEditQuestion))
	// delete question
	puts.Handle("/q/{id}", wrapperHandle(s, s.HandleDeleteQuestion))
	// get vote
	gets.Handle("/q/{id}/vote/{opt}", wrapperHandle(s, s.HandleVote))
	// insert
	posts.Handle("/q/{id}/response",
		wrapperHandle(s, s.HandleQuestionResponse))
	// update
	posts.Handle("/q/{id}/response/{rid}",
		wrapperHandle(s, s.HandleUpdateQuestionResponse))
	// delete
	puts.Handle("/q/{id}/response/{rid}",
		wrapperHandle(s, s.HandleDeleteQuestionResponse))
	// insert response's comment
	posts.Handle("/q/{id}/response/{rid}/comment",
		wrapperHandle(s, s.HandleResponseComment))
	// insert question's comment
	posts.Handle("/q/{id}/comment",
		wrapperHandle(s, s.HandleQuestionComment))
	// update question's comment
	posts.Handle("/q/{id}/comment/{cid}",
		wrapperHandle(s, s.HandleUpdateQuestionComment))
	// delete question's comment
	puts.Handle("/q/{id}/comment/{cid}",
		wrapperHandle(s, s.HandleDeleteQuestionComment))
	gets.Handle("/trendingtags", wrapperHandle(s, s.HandleTrendingTags))
	//s.m.NotFoundHandler = http.HandlerFunc(s.HandleNotFound)
	s.m.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	s.m.NotFoundHandler = wrapperHandle(s, s.HandleNotFound)
	http.Handle("/", s.m)
}

func (s *AEServer) HandleGetDBCount(w http.ResponseWriter,
	r *http.Request) (int, string) {
	q := s.questions.GetCount()/PAGER + 1
	return http.StatusOK, Stringify(q)
}

func (s *AEServer) HandleNotFound(w http.ResponseWriter,
	r *http.Request) (int, string) {
	//w.WriteHeader(http.StatusNotFound)
	//io.WriteString(w, "Not found")
	return http.StatusNotFound, Message("Not Found")
}

func (s *AEServer) HandleRegister(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Registering !")
	var a AuthAttempt
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&a)
	if err != nil {
		log.Println(err)
		return http.StatusNotFound, Message("Register Failed")
	}
	log.Println("Registering new user!")
	log.Println(a)
	user := new(User)
	user.Password = a.Password
	user.Username = a.Username
	user.Salt = a.Salt
	user.ID = bson.NewObjectId()
	err = s.users.Save(user)
	if err != nil {
		log.Println(err)
	}
	return http.StatusOK, Message("Success!")
}

//Get salt associated with a given username
func (s *AEServer) HandleGetSalt(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Get Salt")
	a := AuthFromJson(r.Body)
	if a == nil || a.Username == "" {
		return http.StatusUnauthorized, Message("Must send username!")
	}
	user := s.FindUserByName(a.Username)
	if user == nil {
		return http.StatusUnauthorized, Message("Username not found!")
	}
	salt := genRandString()
	s.salts[a.Username] = salt
	return http.StatusOK, salt
}

func (s *AEServer) HandleLogin(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Login!")
	a := AuthFromJson(r.Body)
	if a == nil {
		time.Sleep(time.Second)
		return http.StatusNotFound, Message("Login Failed")
	}
	sltr := StrResponse{}
	sltr.Arg = a.Username
	sltr.Resp = make(chan string)
	s.ch_getsalt <- sltr
	salt := <-sltr.Resp
	if salt == "" {
		return http.StatusUnauthorized, Message("No login salt registered!")
	}
	user := s.FindUserByName(a.Username)
	salt_pass := DoHash(user.Password, salt)
	if salt_pass != a.Password {
		log.Println("Invalid password.")
		time.Sleep(time.Second)
		return http.StatusUnauthorized, Message("Invalid Username or Password.")
	}
	user.login = time.Now()
	tok := s.GetSessionToken()
	s.ch_login <- &Session{tok, user}
	s.sessionStore.SetParam(w, r)
	s.sessionStore.Put("Login", tok)
	log.Println("Logged in!")
	return http.StatusOK, Stringify(user)
}

func (s *AEServer) HandleUniqueSalt(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle UniqueSalt!")
	a := AuthFromJson(r.Body)
	if a == nil || a.Username == "" {
		return http.StatusUnauthorized, Message("No Username given.")
	}
	user := s.FindUserByName(a.Username)
	if user == nil {
		return http.StatusUnauthorized, Message("No such user!")
	}
	return http.StatusOK, fmt.Sprintf("{\"Salt\":\"%s\"}", user.Salt)
}

func (s *AEServer) HandleLogout(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle logout!")
	s.sessionStore.SetParam(w, r)
	toki, ok := s.sessionStore.Get("Login")
	if !ok {
		//return http.StatusNotFound, Message("something wrong?")
		//return http.StatusOK, Message("need check ok")
		return http.StatusUnauthorized, Message("session expired")
	}
	s.ch_logout <- toki.(string)
	s.sessionStore.Del("Login")
	return http.StatusOK, Message("ok")
}

func (s *AEServer) HandleMe(w http.ResponseWriter,
	r *http.Request) (int, string) {
	return http.StatusOK, Message("Nothing here")
}

func (s *AEServer) HandlePostQuestion(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Post Question!")
	//Verify user account or something
	s.sessionStore.SetParam(w, r)
	login, ok := s.sessionStore.Get("Login")
	if !ok {
		log.Println("Get Key Error - why?")
		//return http.StatusNotFound, Message("something wrong?")
		return http.StatusUnauthorized, Message("session expired")
	}
	tok := login.(string)
	user := s.syncGetUser(tok)
	if user == nil {
		return http.StatusBadRequest, Message("Invalid Cookie!")
	}
	q := QuestionFromJson(r.Body)
	if q == nil {
		return http.StatusNotFound, Message("Poorly Formatted JSON.")
	}
	//Assign question an ID
	q.ID = bson.NewObjectId()
	q.Author = user.Username
	q.Timestamp = time.Now()
	err := s.questions.Save(q)
	if err != nil {
		log.Print(err)
		return http.StatusInternalServerError, Message("Failed to save question")
	}
	return http.StatusOK, q.GetIdHex()
}

func (s *AEServer) HandleGetQuestions(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Get Questions!")
	//vars := mux.Vars(r)
	page := r.FormValue("page")
	if page == "" {
		// set default
		page = "1"
	}
	n, _ := strconv.Atoi(page)
	log.Printf("page number = %d\n", n)
	// db.users.find().skip(pagesize * (n - 1)).limit(pagesize)
	q := s.questions.FindWhereN(bson.M{}, n)
	//q := s.questions.Find(bson.M{}).Skip(5 * (n - 1)).Limit(5)
	if q == nil {
		log.Println("question not found ?")
		return http.StatusNotFound, Message("Question not found.")
	}
	return http.StatusOK, Stringify(q)
}

func (s *AEServer) HandleGetQuestion(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Get Question - with id")
	vars := mux.Vars(r)
	id := vars["id"]
	log.Println("variable id = " + id)
	hid := bson.ObjectIdHex(id)
	log.Println(hid)
	q, ok := s.questions.FindByID(hid).(*Question)
	if !ok || q == nil {
		return http.StatusNotFound, ""
	}
	return http.StatusOK, Stringify(q)
}

func (s *AEServer) HandleQuestionComment(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Question Comment")
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not authorized to reply!")
	}
	comment := user.MakeComment(r.Body)
	if comment == nil {
		return http.StatusBadRequest, Message("Poorly formatted JSON")
	}
	log.Println("Working ...")
	question, ok := s.questions.FindByID(id).(*Question)
	if !ok {
		return http.StatusForbidden, Message("No such question!")
	}
	question.AddComment(comment)
	question.LastEdit = time.Now()
	s.questions.Update(question)
	return http.StatusOK, string(comment.JsonBytes())
}

func (s *AEServer) HandleUpdateQuestionComment(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Update Question Comment")
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	cid := bson.ObjectIdHex(vars["cid"])
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not authorized to reply!")
	}
	comment := user.UpdateComment(r.Body, cid)
	if comment == nil {
		return http.StatusBadRequest, Message("Poorly formatted JSON")
	}
	question, ok := s.questions.FindByID(id).(*Question)
	if !ok {
		return http.StatusForbidden, Message("No such question!")
	}
	question.UpdateComment(comment)
	question.LastEdit = time.Now()
	s.questions.Update(question)
	return http.StatusOK, string(comment.JsonBytes())
}

func (s *AEServer) HandleDeleteQuestionComment(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Delete Question Comment")
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	cid_org := vars["cid"]
	cid := bson.ObjectIdHex(cid_org)
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not authorized to reply!")
	}
	question, ok := s.questions.FindByID(id).(*Question)
	if !ok {
		return http.StatusForbidden, Message("No such question!")
	}
	if question.Author != user.Username {
		return http.StatusForbidden, Message("No such question!")
	}
	question.DeleteComment(cid)
	//question.LastEdit = time.Now()
	s.questions.Update(question)
	//log.Println(fmt.Sprintf(`{"ID":"%s"}`, cid))
	return http.StatusOK, fmt.Sprintf(`{"ID":"%s"}`, cid_org)
}

func (s *AEServer) HandleEditQuestion(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Edit Question")
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not authorized to edit!")
	}
	q := QuestionFromJson(r.Body)
	if q == nil {
		return http.StatusBadRequest, Message("Poorly formatted JSON")
	}
	original := s.questions.FindByID(id).(*Question)
	original.Body = q.Body
	original.Title = q.Title
	original.LastEdit = time.Now()
	s.questions.Update(original)
	return http.StatusOK, Stringify(original)
}

func (s *AEServer) HandleDeleteQuestion(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Delete Question")
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not authorized to edit!")
	}
	original, ok := s.questions.FindByID(id).(*Question)
	if !ok {
		return http.StatusForbidden, Message("No such question!")
	}
	if original.Author != user.Username {
		return http.StatusForbidden, Message("No such question!")
	}
	s.questions.Delete(id)
	return http.StatusOK, Stringify("{}")
}

func (s *AEServer) HandleQuestionResponse(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Question Response")
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not authorized to reply!")
	}
	resp := user.MakeResponse(r.Body)
	if resp == nil {
		return http.StatusBadRequest, Message("Poorly formatted JSON")
	}
	question, ok := s.questions.FindByID(id).(*Question)
	if !ok {
		return http.StatusForbidden, Message("No such question!")
	}
	question.AddResponse(resp)
	question.LastEdit = time.Now()
	s.questions.Update(question)
	return http.StatusOK, Stringify(resp)
}

func (s *AEServer) HandleUpdateQuestionResponse(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Update Question Response")
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	rid := bson.ObjectIdHex(vars["rid"])
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not authorized to reply!")
	}
	resp := user.UpdateResponse(r.Body, rid)
	if resp == nil {
		return http.StatusBadRequest, Message("Poorly formatted JSON")
	}
	question, ok := s.questions.FindByID(id).(*Question)
	if !ok {
		return http.StatusForbidden, Message("No such question!")
	}
	question.UpdateResponse(resp)
	question.LastEdit = time.Now()
	s.questions.Update(question)
	return http.StatusOK, Stringify(resp)
}

// @todo : working more !
func (s *AEServer) HandleDeleteQuestionResponse(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Delete Question Response")
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	rid_org := vars["rid"]
	rid := bson.ObjectIdHex(rid_org)
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not authorized to reply!")
	}
	question, ok := s.questions.FindByID(id).(*Question)
	if !ok {
		return http.StatusForbidden, Message("No such question!")
	}
	resp := question.GetResponse(rid)
	if resp == nil {
		return http.StatusForbidden, Message("No such question!")
	}
	if resp.Author != user.Username {
		return http.StatusForbidden, Message("No such question!")
	}
	question.DeleteResponse(rid)
	//question.LastEdit = time.Now()
	s.questions.Update(question)
	return http.StatusOK, fmt.Sprintf(`{"ID":"%s"}`, rid_org)
}

func (s *AEServer) HandleResponseComment(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Response Comment")
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not authorized to reply!")
	}
	comment := user.MakeComment(r.Body)
	if comment == nil {
		return http.StatusBadRequest, Message("Poorly formatted JSON")
	}
	question, ok := s.questions.FindByID(id).(*Question)
	if !ok {
		return http.StatusForbidden, Message("No such question!")
	}
	rid := vars["rid"]
	resp := question.GetResponse(bson.ObjectId(rid))
	resp.AddComment(comment)
	question.LastEdit = time.Now()
	s.questions.Update(question)
	return http.StatusOK, string(comment.JsonBytes())
}

func (s *AEServer) HandleVote(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Vote")
	vars := mux.Vars(r)
	opt := vars["opt"]
	//opt := params["opt"]
	if opt != "up" && opt != "down" {
		return http.StatusMethodNotAllowed, Message("Invalid vote type")
	}
	user := s.GetAuthedUser(w, r)
	if user == nil {
		return http.StatusUnauthorized, Message("Not logged in!")
	}
	//q := bson.ObjectIdHex(params["id"])
	q := bson.ObjectIdHex(vars["id"])
	question, ok := s.questions.FindByID(q).(*Question)
	if question == nil || !ok {
		return http.StatusNotFound, Message("No such question!")
	}
	question.LastEdit = time.Now()
	switch opt {
	case "up":
		if question.Upvote(user.ID) {
			s.questions.Update(question)
		}
	case "down":
		if question.Downvote(user.ID) {
			s.questions.Update(question)
		}
	}
	return http.StatusOK, Stringify(question)
}

func (s *AEServer) HandleTrendingTags(w http.ResponseWriter,
	r *http.Request) (int, string) {
	log.Println("Handle Trending Tags")
	db_tags := s.questions.FindSelect(bson.M{}, bson.M{"tags": 1})
	//db_tags := s.questions.FindWhere(bson.M{"tags": 1})
	if db_tags == nil {
		log.Println("tags not found ?")
		return http.StatusNotFound, Message("tags not found.")
	}
	map_tags := make(map[string]int)
	for _, v := range db_tags {
		q := v.(*Question)
		for _, vt := range q.Tags {
			count, ok := map_tags[vt]
			if !ok {
				map_tags[vt] = 1
			} else {
				map_tags[vt] = count + 1
			}
		}
		//log.Println(q.Tags)
	}
	var tags_array []string
	var i = 0
	for k, _ := range map_tags {
		tags_array = append(tags_array, k)
		if i += 1; i == 20 {
			break
		}
	}
	return http.StatusOK, Stringify(tags_array)
}
