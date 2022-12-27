package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type TestCollection struct {
	Name  string `json:"name"`
	Vueal string `json:"value"`
}

type TestCollectionID struct {
	Id    string `json:"id" bson:"_id"`
	Name  string `json:"name"`
	Vueal string `json:"value"`
}

func TestMongoCRUD(t *testing.T) {
	driver, err := NewMongoDriver(MongoDSN{
		UserName: "user",
		Password: "123456",
		Host:     "localhost:27017",
		DB:       "mallcook",
	})
	if err != nil {
		t.Fatal(err)
	}

	// 新建集合: 自动创建

	// 新建文档
	insertResult, err := driver.Collection("testcollection").InsertOne(context.TODO(), &TestCollection{Name: "hello", Vueal: "world"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	// 查询文档
	var project TestCollectionID
	result := driver.Collection("testcollection").FindOne(context.TODO(), bson.M{"name": "hello"})
	err = result.Decode(&project)
	if err == mongo.ErrNoDocuments {
		// Do something when no record was found
		fmt.Println("record does not exist")
	} else if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Select a single document: ", project.Id)

	// 转成json
	data, err := driver.Collection("testcollection").FindOne(context.TODO(), bson.M{"name": "hello"}).DecodeBytes()
	if err != nil {
		t.Fatal(err)
	}
	var datajson interface{}
	if err := json.Unmarshal([]byte(data.String()), &datajson); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(datajson)
	}

	// 删除集合
	if err := driver.Collection("testcollection").Drop(context.TODO()); err != nil {
		t.Fatal(err)
	}
}
