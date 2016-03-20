package main

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessoinStorer(t *testing.T) {
	secret, err := ioutil.ReadFile("../secret")
	if err != nil {
		t.Error("Error")
	}
	decode, _ := base64.StdEncoding.DecodeString(string(secret))
	sess := NewSessionStorer("test", decode)
	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp := httptest.NewRecorder()
	sess.SetParam(rsp, req)
	sess.Put("hello", "what")
	val, _ := sess.Get("hello")
	t.Log("value = " + val.(string))
}
