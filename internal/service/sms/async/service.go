package async

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
	"webook/internal/domain"
	"webook/internal/repository"
	"webook/internal/service/sms"
)

// 异步SMS 策略

type Service struct {
	mu  *sync.Mutex
	svc sms.Service
	// 转异步，存储短信请求的repo
	repo   repository.AsyncSmsRepository
	option Options

	//连续发送失败次数
	errCnt int64

	// 连续发送成功次数
	successCnt int64

	// 使用chan和goroutine 通信，退出异步机制
	stop chan int
	//TODO: 日志
}

// Options 异步策略配置
/*
	1. 什么情况下异步
	2. 自动异步策略
*/
type Options struct {
	//外部直接开启异步策略，内部服务不干预，优先级大于 AutoAsync
	Async bool
	// 异步策略由内部决策
	AutoAsync bool
	// 异步情况支持几个goroutine
	GoroutineMax int
	//TODO: 其他配置
}

func NewOptions() Options {
	return Options{
		Async: true,
	}
}

func (o *Options) init() {
	if !o.Async && !o.AutoAsync {
		o.AutoAsync = true
	}
	if o.GoroutineMax <= 0 {
		o.GoroutineMax = 1
	}

	if o.GoroutineMax > 32 {
		o.GoroutineMax = 32
	}

	return
}

func (s *Service) needAsync(errCnt, successCnt int64, stop chan int) bool {
	// 如果外部开启异步策略则直接返回，则不会进行 同步与异步之间的互相转换
	if s.option.Async {
		return true
	}

	if s.option.AutoAsync {
		// TODO: 优化 使用函数式编程，策略根据函数调用，可随时跟换
		// 错误率占比比较大的情况下
		if errCnt >= successCnt/2 {
			return true
		} else if successCnt >= errCnt*3 {
			// 转同步状态
			// 为了避免chan被重复关闭panic，其实大概率不会出现但是保险起见还是检查一下
			_, okk := <-stop
			if !okk {
				log.Println("chan被重复关闭")
				return false
			}

			// 关闭chan 加个锁
			s.mu.Lock()
			close(s.stop)
			s.mu.Unlock()
		}
		// 如果stop是关闭的情况下则refreshAsync 返回同步状态 并且情况之前的状态
		_, ok := <-stop
		// 通道已经关闭了

		if !ok {
			// 返回false 证明其他goroutine 已经修改了 不需要继续了,直接切换到同步

			if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&(s.stop))),
				unsafe.Pointer(&stop),
				unsafe.Pointer(&(s.stop))) {
				return false
			}

			if !atomic.CompareAndSwapInt64(&(s.errCnt), errCnt, 0) {
				return false
			}

			if !atomic.CompareAndSwapInt64(&(s.successCnt), successCnt, 0) {
				return false
			}

		}

	}

	//同步
	return false

}

func (s *Service) StartAsync() {
	for {
		select {
		case <-s.stop:
			log.Println("接受到stop信号 批量关闭协程")
			return
		default:
			s.AsyncSend()
		}
	}
}

func NewService(svc sms.Service, repo repository.AsyncSmsRepository, opt Options) *Service {
	s := &Service{
		svc:        svc,
		repo:       repo,
		option:     opt,
		stop:       make(chan int, 1),
		errCnt:     0,
		successCnt: 0,
		mu:         &sync.Mutex{},
	}

	// 初始化option
	opt.init()

	// 返回的时候默认直接开启异步调度 启动多个go routine
	for i := 0; i < opt.GoroutineMax; i++ {
		go func() {
			s.StartAsync()
		}()
	}

	return s

}

func (s *Service) AsyncSend() {
	//从数据库 或者消息队列中 获取需要异步发送的短信
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	as, err := s.repo.PreemptWaitingSMS(ctx)
	switch err {
	case nil:
		// 执行发送
		// 这个也可以做成配置的
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = s.svc.Send(ctx, as.TplId, as.Args, as.Numbers...)
		if err != nil {
			// errCnt ++
			atomic.AddInt64(&(s.errCnt), 1)
			log.Println("执行异步发送短信失败")
		} else {
			// successCnt ++
			atomic.AddInt64(&(s.successCnt), 1)
		}

		res := err == nil
		// 通知 repository 我这一次的执行结果
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			log.Println("执行异步发送短信成功，但是标记数据库失败")
		}
	//case repository.ErrWaitingSMSNotFound:
	//	// 这个地方很明显就是没有异步发送的
	//	// 睡一秒。这个你可以自己决定
	//	time.Sleep(time.Second)
	default:
		// 正常来说应该是数据库那边出了问题，
		// 但是为了尽量运行，还是要继续的
		// 你可以稍微睡眠，也可以不睡眠
		// 睡眠的话可以帮你规避掉短时间的网络抖动问题
		log.Println("抢占异步发送短信任务失败")
		time.Sleep(time.Second)
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 停止异步的chan
	stop := s.stop
	// err次数
	errCnt := s.errCnt
	// 成功次数
	sucCnt := s.successCnt
	if s.needAsync(errCnt, sucCnt, stop) {

		// 如果需要异步发送则直接插入数据库，等待异步调度结果
		err := s.repo.Add(ctx, domain.AsyncSms{
			TplId:    tplId,
			Args:     args,
			Numbers:  numbers,
			RetryMax: 3,
		})
		return err
	}
	return s.svc.Send(ctx, tplId, args, numbers...)
}
