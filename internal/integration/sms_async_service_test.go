package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"webook/internal/integration/startup"
	"webook/internal/service/sms/async"
	"webook/ioc"
)

// 集成测试不使用mock，直接发送真的短信

type AsyncSMSTestSuite struct {
	suite.Suite
	db       *gorm.DB
	asyncSvc *async.Service
}

func (s *AsyncSMSTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE table `async_sms`")
}

func (s *AsyncSMSTestSuite) SetupSuite() {
	s.db = startup.InitDB()
	svc := ioc.InitSMSService()
	asyncService := startup.InitAsyncSmsService(svc)
	s.asyncSvc = asyncService
}

func (s *AsyncSMSTestSuite) TestSend() {
	// 拿到测试testing
	t := s.T()
	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after   func(t *testing.T)
		tplId   string
		args    []string
		numbers []string
		wantErr error
	}{
		{
			name: "异步发送成功",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.asyncSvc.Send(context.Background(), "", []string{"123456"}, "17674123135")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
