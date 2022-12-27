package mongodb

import (
	"context"
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

type MongoProject struct {
	Pages  []interface{} `json:"pages"`
	Cover  string        `json:"cover"`
	Config interface{}   `json:"config"`
	Type   string        `json:"type"`
	UserID string        `json:"userId"`
}

type MongoUser struct {
	ID       string `json:"id" bson:"_id"`
	UserName string `json:"userName"`
}

func (MongoProject) CollectionName() string {
	return "project"
}

func (MongoUser) CollectionName() string {
	return "user"
}

func (u MongoProject) FindByUserID(userID string) (interface{}, error) {
	data, err := Mongo().Collection(u.CollectionName()).FindOne(context.TODO(), bson.M{"userId": userID}).DecodeBytes()
	if err != nil {
		return nil, err
	}

	if result, err := Mongo().DecodeToJSON(data); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (u MongoProject) FindByUserName(userName string) (interface{}, error) {
	var user MongoUser
	if err := Mongo().Collection(MongoUser{}.CollectionName()).FindOne(context.TODO(), bson.M{"userName": userName}).Decode(&user); err != nil {
		return nil, fmt.Errorf("用户查询失败: %w", err)
	}

	data, err := Mongo().Collection(u.CollectionName()).FindOne(context.TODO(), bson.M{"userId": user.ID}).DecodeBytes()
	if err != nil {
		return nil, err
	}

	if result, err := Mongo().DecodeToJSON(data); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}
