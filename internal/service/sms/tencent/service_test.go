package tencent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"testing"
)

// 这个需要手动跑，也就是你需要在本地搞好这些环境变量
func TestSender(t *testing.T) {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		t.Fatal()
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")

	secretId = "AKIDDIAnGcy7FIA752rtnaGk4GVhmzEWglMz"
	secretKey = "Xr3IGD0c2GKQNIJHLPL9oR4shOrCc7Y3"

	c, err := sms.NewClient(common.NewCredential(secretId, secretKey),
		"ap-guangzhou",
		profile.NewClientProfile())
	if err != nil {
		t.Fatal(err)
	}

	s := NewService(c, "1400351467", "张志明的博客")

	testCases := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name:   "发送验证码",
			tplId:  "577583",
			params: []string{"123456", "5"},
			// 改成你的手机号码
			numbers: []string{"17674123135"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			er := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, er)
		})
	}
}
