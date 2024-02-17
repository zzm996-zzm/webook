package domain

type AsyncSms struct {
	Id int64

	// 发生短信的参数
	TplId   string
	Args    []string
	Numbers []string
	// 最大重试次数

	RetryMax int
}
