package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
)

//任务的执行时间点
type TimePoint struct {
	StartTime int64		`bson:"startTime"`
	EndTime int64		`bson:"endTime"`
}

//日志
type LogRecord struct {
	JobName string		`bson:"jobName"`
	Command string		`bson:"command"`
	Err string			`bson:"err"`
	Content string		`bson:"content"`
	TimePoint TimePoint	`bson:"timePoint"`
}

func main() {
	var(
		client *mongo.Client
		err error
		database *mongo.Database
		collection *mongo.Collection
		record *LogRecord
		logArr []interface{}
		insertId interface{}
		result *mongo.InsertManyResult
		docId objectid.ObjectID
	)

	//1.创建连接
	if client, err = mongo.Connect(context.TODO(), "mongodb://47.99.240.52:27017", clientopt.ConnectTimeout(5*time.Second)); err != nil {
		fmt.Println(err)
		return
	}

	//2.选择数据库
	database = client.Database("cron")

	//3.选择表
	collection = database.Collection("log")

	//4.插入记录(bson)
	record = &LogRecord{
		JobName:"job10",
		Command:"echo hello",
		Err:"",
		Content:"hello",
		TimePoint:TimePoint{
			StartTime:time.Now().Unix(),
			EndTime:time.Now().Unix() + 10,
		},
	}

	//5.批量插入多条document
	logArr = []interface{}{record,record,record}

	//发起插入
	if result,err = collection.InsertMany(context.TODO(),logArr);err != nil {
		fmt.Println(err)
		return
	}

	for _,insertId = range result.InsertedIDs{
		//拿着interface{}，反射成objectID
		docId = insertId.(objectid.ObjectID)
		fmt.Println("自增ID:",docId)
	}
}
