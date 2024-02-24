package web

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

type Article struct {
	Id       int64  `bson:"id,omitempty"`
	Title    string `bson:"title,omitempty"`
	Content  string `bson:"content,omitempty"`
	AuthorId int64  `bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}

func TestMongoDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 添加监控
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
	}

	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(monitor)
	// 初始化客户端
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)

	col := client.Database("webook").Collection("articles")

	// 插入的时候可以直接写结构体
	insertRes, err := col.InsertOne(ctx, Article{
		Id:       1,
		Title:    "我的标题",
		Content:  "我的内容",
		AuthorId: 123,
	})

	assert.NoError(t, err)

	oid := insertRes.InsertedID.(primitive.ObjectID)
	t.Log("插入ID", oid)

	// filter
	filter := bson.D{bson.E{Key: "id", Value: 1}}

	findRes := col.FindOne(ctx, filter)
	if findRes.Err() == mongo.ErrNoDocuments {
		t.Log("没找到数据")
	} else {
		assert.NoError(t, findRes.Err())
		var art Article
		err = findRes.Decode(&art)
		assert.NoError(t, err)
		t.Log(art)
	}

	updateFilter := bson.D{bson.E{Key: "id", Value: 1}}
	set := bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: "title", Value: "新的标题"}}}}
	updateOneRes, err := col.UpdateOne(ctx, updateFilter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateOneRes.ModifiedCount)

	//updateMany
	set = bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: "content", Value: "新的内容"}}}}
	updateManyRes, err := col.UpdateMany(ctx, updateFilter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)

	//delete
	deleteFilter := bson.D{bson.E{"id", 1}}
	delRes, err := col.DeleteOne(ctx, deleteFilter)
	assert.NoError(t, err)
	t.Log("删除文档数量", delRes)
}

// 测试套件
type MongoDBTestSuite struct {
	suite.Suite
	col *mongo.Collection
}

func TestMongoDBQueries(t *testing.T) {
	suite.Run(t, &MongoDBTestSuite{})
}

func (s *MongoDBTestSuite) SetupSuite() {
	t := s.T()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 添加监控
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
	}

	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(monitor)
	// 初始化客户端
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)

	col := client.Database("webook").Collection("articles")
	s.col = col

	manyRes, err := col.InsertMany(ctx, []any{
		Article{
			Id:       123,
			Content:  "我的内容123123",
			Title:    "我的标题123123",
			AuthorId: 11,
		},
		Article{
			Id:       123,
			Content:  "我的内容123",
			Title:    "我的标题123",
			AuthorId: 11,
		}, Article{
			Id:       234,
			Content:  "我的内容234",
			Title:    "我的标题234",
			AuthorId: 12,
		}})

	assert.NoError(s.T(), err)
	s.T().Log("插入数量", len(manyRes.InsertedIDs))
}

func (s *MongoDBTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := s.col.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.col.Indexes().DropAll(ctx)
	assert.NoError(s.T(), err)
}

func (s *MongoDBTestSuite) TestOr() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.A{bson.D{bson.E{"id", 123}}, bson.D{bson.E{"id", 234}}}
	res, err := s.col.Find(ctx, bson.D{bson.E{"$or", filter}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log(arts)

}

func (s *MongoDBTestSuite) TestAnd() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.A{bson.D{bson.E{"id", 123}}, bson.D{bson.E{"author_id", 11}}}
	res, err := s.col.Find(ctx, bson.D{bson.E{"$and", filter}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log(arts)
}

func (s *MongoDBTestSuite) TestIn() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.D{bson.E{"id", bson.D{bson.E{"$in", []int{123, 234}}}}}
	//只查询id 的写法
	proj := bson.M{"id": 1}
	res, err := s.col.Find(ctx, filter, options.Find().SetProjection(proj))
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log(arts)
}

func (s *MongoDBTestSuite) TestPage() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 根据 author_id 查询出当前用户的文章数据 分页返回
	var res []Article
	var findOptions options.FindOptions
	findOptions.SetLimit(int64(2))
	findOptions.SetSkip(int64(0))

	filter := bson.D{bson.E{"author_id", 11}}

	find, err := s.col.Find(ctx, filter, &findOptions)

	assert.NoError(s.T(), err)

	err = find.All(ctx, &res)

	assert.NoError(s.T(), err)

	s.T().Log("查询出来的结果：", res)

}

func (s *MongoDBTestSuite) TestIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ires, err := s.col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{bson.E{"id", 1}},
		Options: options.Index().SetUnique(true).SetName("idx_id"),
	})
	assert.NoError(s.T(), err)
	s.T().Log("创建索引", ires)
}
