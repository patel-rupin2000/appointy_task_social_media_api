package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// func Compare(str1, str2 string)int

func TestUserGetEntry(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	req, err := http.NewRequest("GET", "/users/616175759c80294c8458e328", nil)
	if err != nil {
		t.Fatal(err)
	}
	q := req.URL.Query()
	q.Add("_id", "61618f8b9c80294c8458e329")
	req.URL.RawQuery = q.Encode()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetPersonEndpoint)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := []byte(`{"_id":"61618f8b9c80294c8458e329","name":"Rupin Patel","email":"patelrupin63@gmail.com","password":"$2a$14$3f2ftvqEYOLpeF./C4SRW..XqbON3b9FPzmpBArfV6RwPNGjIqG9i"}`)
	expected_string := `{"_id":"61618f8b9c80294c8458e329","name":"Rupin Patel","email":"patelrupin63@gmail.com","password":"$2a$14$3f2ftvqEYOLpeF./C4SRW..XqbON3b9FPzmpBArfV6RwPNGjIqG9i"}`

	ans_bytes := rr.Body.Bytes()
	eq, err := JSONBytesEqual(ans_bytes, expected)

	if !eq {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected_string)
	}

}
