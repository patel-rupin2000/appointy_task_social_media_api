package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func JSONBytesEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}
func TestAllPostGetEntry(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	req, err := http.NewRequest("GET", "/posts/users/61618f8b9c80294c8458e329", nil)
	if err != nil {
		t.Fatal(err)
	}
	q := req.URL.Query()
	q.Add("_id", "61618f8b9c80294c8458e329")
	req.URL.RawQuery = q.Encode()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetPostsEndpoint)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := []byte(`[
		{
			"_id": "616175759c80294c8458e328",
			"caption": "Wow! what a day",
			"imageurl": "http://good_day.jpg",
			"TimeStamp": "Mon Jan  1"
		},
		{
			"_id": "616192249c80294c8458e32a",
			"TimeStamp": "Mon Jan  1"
		}
	]`)
	expected_string := `[
		{
			"_id": "616175759c80294c8458e328",
			"caption": "Wow! what a day",
			"imageurl": "http://good_day.jpg",
			"TimeStamp": "Mon Jan  1"
		},
		{
			"_id": "616192249c80294c8458e32a",
			"TimeStamp": "Mon Jan  1"
		}
	]`

	ans_bytes := rr.Body.Bytes()
	eq, err := JSONBytesEqual(ans_bytes, expected)

	if !eq {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected_string)
	}

}
