package main

import (
	"errors"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var ErrorNotFound = errors.New("No documents found!")
var ErrorNullResponse = errors.New("Got back null response from mgo.")

const PAGER int = 5

type Database struct {
	db          *mgo.Database
	collections map[string]*Collection
}

func NewDatabase(host string) *Database {
	log.Println("NewDatabase ....." + host)
	s, err := mgo.Dial(host)
	if err != nil {
		panic(err)
	}
	mdb := s.DB("askeecs")

	dbs := new(Database)
	dbs.db = mdb
	dbs.collections = make(map[string]*Collection)
	return dbs
}

func (db *Database) Collection(name string, typ I) *Collection {
	c, ok := db.collections[name]
	if ok {
		return c
	}

	c = new(Collection)
	c.col = db.db.C(name)
	c.template = typ
	return c
}

type I interface {
	GetID() bson.ObjectId
	New() I
}

type Collection struct {
	col      *mgo.Collection
	cache    map[string]I
	template I
}

func (c *Collection) Save(doc I) error {
	//TODO: handle errors?
	log.Printf("Saving document.")
	err := c.col.Insert(doc)
	return err
}

func (c *Collection) Update(doc I) error {
	log.Println("Updating Document.")
	err := c.col.UpdateId(doc.GetID(), doc)
	return err
}

func (c *Collection) FindByID(ID bson.ObjectId) I {
	q := c.col.FindId(ID)
	if q == nil {
		log.Println(ErrorNullResponse)
		return nil
	}
	cnt, err := q.Count()
	if err != nil {
		log.Println(err)
		return nil
	}
	if cnt < 1 {
		log.Println(ErrorNotFound)
		return nil
	}
	obj := c.template.New()
	q.One(obj)
	return obj
}

func (c *Collection) FindSelect(match bson.M, sel bson.M) []I {
	log.Println(match)
	q := c.col.Find(match).Select(sel)
	if q == nil {
		log.Println(ErrorNullResponse)
		return nil
	}

	n, err := q.Count()
	if err != nil {
		log.Println(err)
		return nil
	}
	if n == 0 {
		log.Println("Nothing matched the query...")
		return nil
	}

	var out []I
	i := q.Iter()
	v := c.template.New()
	for i.Next(v) {
		out = append(out, v)
		v = c.template.New()
	}
	return out
}

func (c *Collection) FindWhere(match bson.M) []I {
	log.Println(match)
	q := c.col.Find(match)
	if q == nil {
		log.Println(ErrorNullResponse)
		return nil
	}

	n, err := q.Count()
	if err != nil {
		log.Println(err)
		return nil
	}
	if n == 0 {
		log.Println("Nothing matched the query...")
		return nil
	}

	var out []I
	i := q.Iter()
	v := c.template.New()
	for i.Next(v) {
		out = append(out, v)
		v = c.template.New()
	}
	return out
}

func (c *Collection) FindWhereN(match bson.M, page int) []I {
	log.Println(match)
	//total := c.GetCount()
	limit := PAGER
	//skip := total - PAGER*page
	skip := PAGER * (page - 1)
	if skip < 0 {
		skip = 0
	}
	q := c.col.Find(match).Sort("-timestamp").Skip(skip).Limit(limit)
	if q == nil {
		log.Println(ErrorNullResponse)
		return nil
	}

	n, err := q.Count()
	if err != nil {
		log.Println(err)
		return nil
	}
	if n == 0 {
		log.Println("Nothing matched the query...")
		return nil
	}

	var out []I
	i := q.Iter()
	v := c.template.New()
	for i.Next(v) {
		out = append(out, v)
		v = c.template.New()
	}
	return out
}

func (c *Collection) GetCount() int {
	log.Println("db - GetCount called")
	//q := c.col.Find(bson.M{})
	n, err := c.col.Count()
	if err != nil {
		log.Println(err)
		return 0
	}
	return n
}

func (c *Collection) Delete(ID bson.ObjectId) error {
	log.Println("db - Delete called")
	err := c.col.RemoveId(ID)
	if err != nil {
		log.Println("db Not Found ID")
		return err
	}

	return nil
}
