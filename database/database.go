package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Song struct {
	SongId            string `bson:"songId"`
	Name              string `bson:"name"`
	Artist            string `bson:"artist"`
	Cover             string `bson:"cover"`
	SpotifyId         string `bson:"spotifyId"`
	NumClicksSongId   int    `bson:"numClicks-songId"`
	NumClickSpotifyId int    `bson:"numClicks-spotifyId"`
}

type db struct {
	collection *mongo.Collection
}

type DB interface {
	InsertSong(song *Song)
	GetSong(song string) (*Song, error)
	AddClick(id string, target string)
}

func NewDB(collection *mongo.Collection) DB {
	db := &db{
		collection: collection,
	}

	return db
}

func CollectionFactory(databaseName string, collectionName string) *mongo.Collection {
	mCl, err := initDB()
	if err != nil {
		return nil
	}
	return mCl.Database(databaseName).Collection(collectionName)
}

func initDB() (*mongo.Client, error) {
	fmt.Println("Initializing the DB")
	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017/")
	mongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = mongoClient.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connection to MongoDB established!")

	// hometown := &Song{
	// 	Name:      "Hometown",
	// 	SongId:    "hometown",
	// 	SpotifyId: "47x1Gh7yk5mblUWxWRdtjH",
	// }
	// insertSong(mongoClient, hometown)

	return mongoClient, nil
}

func (db *db) InsertSong(song *Song) {
	song.NumClickSpotifyId = 0
	result, err := db.collection.InsertOne(context.TODO(), song)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inserted %s with ID: %s. MongoID: %s\n", song.Name, song.SongId, result.InsertedID)
}

func (db *db) GetSong(song string) (*Song, error) {
	filter := bson.D{{Key: "songId", Value: song}}
	var result *Song
	err := db.collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *db) AddClick(id string, target string) {
	filter := bson.D{{Key: target, Value: id}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "numClicks-" + target, Value: 1}}}}
	_, err := db.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Printf("Register click failed. Target: %s\n", target)
	}
}
