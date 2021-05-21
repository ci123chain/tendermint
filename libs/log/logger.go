package log

import (
	"fmt"
	"io"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"gitlab.oneitfarm.com/bifrost/cilog"
	"gitlab.oneitfarm.com/bifrost/cilog/redis_hook"
	logger "gitlab.oneitfarm.com/bifrost/cilog/v2"
	"go.uber.org/zap/zapcore"
)

// Logger is what any Tendermint library should take.
type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	With(keyvals ...interface{}) Logger
}

// NewSyncWriter returns a new writer that is safe for concurrent use by
// multiple goroutines. Writes to the returned writer are passed on to w. If
// another write is already in progress, the calling goroutine blocks until
// the writer is available.
//
// If w implements the following interface, so does the returned writer.
//
//    interface {
//        Fd() uintptr
//    }
func NewSyncWriter(w io.Writer) io.Writer {
	InitOneitfarmLogger()
	return kitlog.NewSyncWriter(w)
}

var LogModeOneitfarm = false

func OneitfarmLog() bool {
	return LogModeOneitfarm
}

func InitOneitfarmLogger() {
	appID := os.Getenv("IDG_APPID")
	if len(appID) == 0 {
		return
	}
	LogModeOneitfarm = true

	conf := &logger.Conf{
		Level:  zapcore.InfoLevel, // 输出日志等级
		Caller: true,              // 是否开启记录调用文件夹+行数+函数名
		// 输出到 redis 的日志全部都是 info 级别以上
		// 不用填写 AppName, AppID 默认会从环境变量获取
		AppInfo: &cilog.ConfigAppData{
			AppVersion: os.Getenv("IDG_VERSION"),
			Language:   "zh-cn",
		},
		HookConfig: &redis_hook.HookConfig{
			Key:  "service_" + appID,            // 填写日志 key
			Host: "redis-cluster-proxy-log.msp", // 填写 log proxy host
			// k8s 集群内填写 redis-cluster-proxy-log.msp
			Port: 6380, // 填写 log proxy port
			// 默认填写 6380
		},
	}

	err := logger.GlobalConfig(*conf)
	if err != nil {
		// 处理 logger 初始化错误
		// log-proxy 连接失败会报错
		// 若不影响程序执行，可忽视
		panic(fmt.Sprintf("[ERR] Logger init error: %v", err))
	}
}
