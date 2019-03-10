package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
	"time"
)

//执行时间点
type TimePoint struct {
	StartTime int64		`bson:"startTime"`
	EndTime int64		`bson:"endTime"`
}

//日志
type LogRecord struct {
	JobName string			`bson:"jobName"`
	Command string			`bson:"command"`
	Err string				`bson:"err"`
	Content string			`bson:"content"`
	TimePoint TimePoint		`bson:"timePoint"`
}


//jobName过滤条件
type FindByJobName struct {
	JobName string	`bson:"jobName"`	//jobName赋值为10
}


func main() {
	var(
		client *mongo.Client
		err error
		database *mongo.Database
		collection *mongo.Collection
		record *LogRecord
		cond *FindByJobName
		cursor mongo.Cursor
	)

	//1.连接mongodb
	if client,err = mongo.Connect(context.TODO(),"mongodb://47.99.240.52:27017",clientopt.ConnectTimeout(5 * time.Second));err != nil {
		fmt.Println(err)
		return
	}

	//2.选择数据库
	database = client.Database("cron")

	//3.选择表
	collection = database.Collection("log")

	//4.按照jobName字段过滤，想找出jobName = 10.找出5条
	cond = &FindByJobName{JobName:"job10"}		//{"jobName:"job10""}

	//5.查询（过滤 + 翻页参数）
	if cursor, err = collection.Find(context.TODO(), cond, findopt.Skip(0), findopt.Limit(2)); err != nil {
		fmt.Println(err)
		return
	}

	//延迟释放游标
	defer cursor.Close(context.TODO())

	//6.遍历结果集
	for cursor.Next(context.TODO()){
		//定义一个日志对象
		record = &LogRecord{}

		//反序列化bson对象
		if err = cursor.Decode(record); err != nil {
			fmt.Println(err)
			return
		}

		// 把日志行打印出来
		fmt.Println(*record)
	}
}
