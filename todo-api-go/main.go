package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    uri := "mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.5.2"
    docs := "www.mongodb.com/docs/drivers/go/current/"
    if uri == "" {
        log.Fatal("Set your 'MONGODB_URI' environment variable. " +
            "See: " + docs +
            "usage-examples/#environment-variable")
    }

    client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
    if err != nil {
        log.Fatal(err)
    }

    defer func() {
        if err := client.Disconnect(context.TODO()); err != nil {
            log.Fatal(err)
        }
    }()

    // Change to your actual database — seems like it's "test"
    coll := client.Database("test").Collection("movies")

    var result bson.M
    filter := bson.D{{"Name", "tiger"}} // match field exactly

    err = coll.FindOne(context.TODO(), filter).Decode(&result)
    if err == mongo.ErrNoDocuments {
        fmt.Println("No document was found with the name tiger")
        return
    }
    if err != nil {
        log.Fatal(err)
    }

    jsonData, err := json.MarshalIndent(result, "", "    ")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s\n", jsonData)
}
