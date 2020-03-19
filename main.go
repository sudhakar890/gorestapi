package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var coll *mongo.Collection
var ctx context.Context // Dont define the conetxt as global variable as it will result in timeouts

// Person model
type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

func createPerson(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var person Person
	json.NewDecoder(request.Body).Decode(&person)
	result, err := coll.InsertOne(ctx, person)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(result)
}

func getPeople(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var people []Person
	cursor, err := coll.Find(ctx, Person{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)
	}
	if err = cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(people)
}

func getPerson(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&person)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(person)
}

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://<user>:<password>@cluster0-9tn5d.azure.mongodb.net/test?retryWrites=true&w=majority"))
	if err != nil {
		log.Panic(err)
	}
	// Context for timeout of 10 secs
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Panic(err)
	}
	defer client.Disconnect(ctx)
	db = client.Database("polyglotdeveloper")
	coll = db.Collection("people")

	router := mux.NewRouter()
	//Endpoints
	router.HandleFunc("/person", createPerson).Methods("POST")
	router.HandleFunc("/people", getPeople).Methods("GET")
	router.HandleFunc("/person/{id}", getPerson).Methods("GET")

	http.ListenAndServe(":8080", router)
}
