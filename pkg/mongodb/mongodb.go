package mongodb

// 基于mongo-driver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDSN struct {
	UserName string
	Password string
	Host     string
	DB       string
}

type MongoDriver struct {
	Client *mongo.Client

	DB string
}

type BsonDecode struct {
	result interface{}
}

// 将_开头的删除
func (b *BsonDecode) UnmarshalJSON(data []byte) error {
	var tmp map[string]interface{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	result := map[string]interface{}{}
	for key, val := range tmp {
		if len(key) == 0 || key[0] == '_' {
			continue
		}
		result[key] = val
	}
	b.result = result
	return nil
}

func (m MongoDSN) String() string {
	return fmt.Sprintf("mongodb://%s:%s@%s/%s", m.UserName, m.Password, m.Host, m.DB)
}

// dsn: mongodb://user:123456@127.0.0.1:27017/mallcook
func NewMongoDriver(dsn MongoDSN) (*MongoDriver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn.String()))
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return &MongoDriver{
		Client: client,
		DB:     dsn.DB,
	}, nil
}

func (d MongoDriver) Collection(collectionName string) *mongo.Collection {
	return d.Client.Database(d.DB).Collection(collectionName)
}

func (d *MongoDriver) Close() error {
	return d.Client.Disconnect(context.TODO())
}

func (d *MongoDriver) DecodeToJSON(raw bson.Raw) (interface{}, error) {
	var result BsonDecode
	if err := json.Unmarshal([]byte(raw.String()), &result); err != nil {
		return nil, err
	} else {
		return result.result, nil
	}
}

// func (d *MongoDriver) InsertOne(collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
// 	collection := d.Client.Database(d.DB).Collection(collectionName)
// 	return collection.InsertOne(context.TODO(), document)
// }

// // 插入多个
// func (d *MongoDriver) InsertMany(collectionName string, documents []interface{}) (*mongo.InsertManyResult, error) {
// 	collection := d.Client.Database(d.DB).Collection(collectionName)
// 	return collection.InsertMany(context.TODO(), documents)
// }

// 查询单个
// func (d *MongoDriver) FindOne(collectionName string, key string, value interface{}) *mongo.SingleResult {
// 	collection := d.Client.Database(d.DB).Collection(collectionName)

// 	filter := bson.M{{Name: key, Value: value}}
// 	return collection.FindOne(context.TODO(), filter)
// }

// 获取集合创建时间和编号
// func (d *MongoDriver) ParsingId(result string) (time.Time, uint64) {
// 	temp1 := result[:8]
// 	timestamp, _ := strconv.ParseInt(temp1, 16, 64)
// 	dateTime := time.Unix(timestamp, 0) // 这是截获情报时间 时间格式 2019-04-24 09:23:39 +0800 CST
// 	temp2 := result[18:]
// 	count, _ := strconv.ParseUint(temp2, 16, 64) // 截获情报的编号
// 	return dateTime, count
// }

// // 按选项查询集合
// // Skip 跳过
// // Limit 读取数量
// // Sort  排序   1 倒叙 ， -1 正序
// func (d *MongoDriver) FindAll(collectionName string, Skip, Limit int64, sort int) (*mongo.Cursor, error) {
// 	collection := d.Client.Database(d.DB).Collection(collectionName)

// 	SORT := bson.D{{Name: "_id", Value: sort}}
// 	filter := bson.D{{}}

// 	// where
// 	findOptions := options.Find()
// 	findOptions.SetSort(SORT)
// 	findOptions.SetLimit(Limit)
// 	findOptions.SetSkip(Skip)

// 	return collection.Find(context.TODO(), filter, findOptions)
// }

// 查询count总数
// func (d *MongoDriver) Count() (string, int64) {
// 	name := m.collection.Name()
// 	size, _ := m.collection.EstimatedDocumentCount(context.TODO())
// 	return name, size
// }

// // 删除
// func (m *Mgo) Delete(key string, value interface{}) int64 {
// 	filter := bson.D{{key, value}}
// 	count, err := m.collection.DeleteOne(context.TODO(), filter, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return count.DeletedCount

// }

// // 删除多个
// func (m *Mgo) DeleteMany(key string, value interface{}) int64 {
// 	filter := bson.D{{key, value}}
// 	count, err := m.collection.DeleteMany(context.TODO(), filter)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return count.DeletedCount
// }

// 更新一个
// func (m *Mgo) UpdateOne(key string, value interface{}, update interface{}) (updateResult *mongo.UpdateResult) {
// 	filter := bson.D{{key, value}}
// 	updateResult, err := m.collection.UpdateOne(context.TODO(), filter, update)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return updateResult
// }
