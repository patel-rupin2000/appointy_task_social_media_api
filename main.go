package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"net/mail"
	"sync"
	"time"

	. "github.com/gobeam/mongo-go-pagination"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func valid_email(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

type Marshaler interface {
	MarshalJSON() ([]byte, error)
}
type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {

	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("Mon Jan _2"))
	return []byte(stamp), nil
}

type Person struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}
type Post struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Caption   string             `json:"caption,omitempty" bson:"caption,omitempty"`
	ImageURL  string             `json:"imageurl,omitempty" bson:"imageurl,omitempty"`
	TimeStamp JSONTime
}

var client *mongo.Client
var lock sync.Mutex

func CreatePersonEndpoint(response http.ResponseWriter, request *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	var limit int64 = 10
	var page int64 = 1
	response.Header().Add("content-type", "application/json")
	var person Person

	json.NewDecoder(request.Body).Decode(&person)
	fmt.Println(hash(person.Email + person.Password))
	//checking for any duplicate or overlaping of password or email ids prevent users to not to use same email ids and password
	collection := client.Database("rupin_patel_appointy").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	match := bson.M{"$match": bson.M{"email": bson.M{"$gt": person.Email}}}

	//group query
	projectQuery := bson.M{"$project": bson.M{"_id": 1, "email": person.Email}}
	aggPaginatedData, err := New(collection).Context(ctx).Limit(limit).Page(page).Aggregate(match, projectQuery)
	if err != nil {
		panic(err)
	}
	var people []Person
	for _, raw := range aggPaginatedData.Data {
		var peoples *Person
		if marshallErr := bson.Unmarshal(raw, &peoples); marshallErr == nil {
			people = append(people, *peoples)
		}

	}

	fmt.Printf("Aggregate Pagination Data for similar email_ids: %+v\n", people)
	if !valid_email(person.Email) {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte("invalid email id please try again"))
		return

	}
	cursor, err := collection.Find(ctx, bson.M{})
	for cursor.Next(ctx) {
		var backlogperson Person
		cursor.Decode(&backlogperson)

		match := CheckPasswordHash(person.Password, backlogperson.Password)
		fmt.Println("Match:   ", match)
		if match == true || backlogperson.Email == person.Email {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{"message password or email is already been used":"` + err.Error() + `"}`))
			return
		}

	}
	hash_password, _ := HashPassword(person.Password)
	person.Password = hash_password

	result, _ := collection.InsertOne(ctx, person)
	json.NewEncoder(response).Encode(result)
	fmt.Println(result.InsertedID)
	time.Sleep(1 * time.Second)

}
func GetPeopleEndpoint(response http.ResponseWriter, request *http.Request) {
	lock.Lock()
	defer lock.Unlock()

	response.Header().Add("content-type", "application/json")
	var people []Person
	collection := client.Database("rupin_patel_appointy").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)

	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(people)
	time.Sleep(1 * time.Second)

}
func GetPersonEndpoint(response http.ResponseWriter, request *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	collection := client.Database("rupin_patel_appointy").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(person)
	time.Sleep(1 * time.Second)

}
func CreatePostEndpoint(response http.ResponseWriter, request *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	var limit int64 = 10
	var page int64 = 1
	response.Header().Add("content-type", "application/json")
	var post Post
	json.NewDecoder(request.Body).Decode(&post)
	collection := client.Database("rupin_patel_appointy").Collection("post")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	match := bson.M{"$match": bson.M{"caption": bson.M{"$gt": post.Caption}}}

	//group query of similar posts
	projectQuery := bson.M{"$project": bson.M{"_id": 1, "caption": post.Caption}}
	aggPaginatedData, err := New(collection).Context(ctx).Limit(limit).Page(page).Aggregate(match, projectQuery)
	if err != nil {
		panic(err)
	}
	var all_posts []Post
	for _, raw := range aggPaginatedData.Data {
		var posts *Post
		if marshallErr := bson.Unmarshal(raw, &posts); marshallErr == nil {
			all_posts = append(all_posts, *posts)
		}

	}
	fmt.Printf("Aggregate Posts List of same caption: %+v\n", all_posts)

	result, _ := collection.InsertOne(ctx, post)
	json.NewEncoder(response).Encode(result)
	time.Sleep(1 * time.Second)
}
func GetPostEndpoint(response http.ResponseWriter, request *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var post Post
	collection := client.Database("rupin_patel_appointy").Collection("post")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, Post{ID: id}).Decode(&post)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(post)
	time.Sleep(1 * time.Second)

}
func GetPostsEndpoint(response http.ResponseWriter, request *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	fmt.Println(id)
	var posts []Post
	collection := client.Database("rupin_patel_appointy").Collection("post")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var post Post
		cursor.Decode(&post)
		posts = append(posts, post)

	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(posts)
	time.Sleep(1 * time.Second)
}

func main() {
	fmt.Println("Starting the Applications")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/users", CreatePersonEndpoint).Methods("POST")
	router.HandleFunc("/users_info", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/users/{id}", GetPersonEndpoint).Methods("GET")
	router.HandleFunc("/posts", CreatePostEndpoint).Methods("POST")
	router.HandleFunc("/posts/{id}", GetPostEndpoint).Methods("GET")
	router.HandleFunc("/posts/users/{id}", GetPostsEndpoint).Methods("GET")

	http.ListenAndServe(":12345", router)

}
