package master

import (
	"context"
	"github.com/mongodb/mongo-go-driver/mongo"
	"go-corntab/common"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

//mongodb日志管理
type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

func InitLogMgr() (err error) {
	var (
		client *mongo.Client
	)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(G_config.MongodbConnectTimeout)*time.Millisecond)
	defer cancel()
	//建立mongodb连接
	if client, err = mongo.Connect(ctx, options.Client().ApplyURI(G_config.MongodbUri)); err != nil {
		return
	}

	G_logMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}
	return
}

//查看任务日志
func (logMgr *LogMgr) ListLog(name string, skip int, limit int) (logArr []*common.JobLog, err error) {
	var (
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor  *mongo.Cursor
		jobLog  *common.JobLog
	)

	logArr = make([]*common.JobLog, 0)

	//过滤条件
	filter = &common.JobLogFilter{JobName: name}

	//按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{SortOrder: -1}

	var skip1 = int64(skip)
	var limit1 = int64(limit)
	//查询
	cursor, err = logMgr.logCollection.Find(context.TODO(), filter, &options.FindOptions{
		Sort:  logSort,
		Skip:  &skip1,
		Limit: &limit1,
	})
	if err != nil {
		return
	}

	//延迟释放游标
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}

		//反序列化BSON
		if err = cursor.Decode(jobLog); err != nil {
			continue //有日志不合法
		}

		logArr = append(logArr, jobLog)
	}

	return
}
