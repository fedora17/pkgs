package dotask

import (
	"pkgs/rotatelog"
)

func RegisterBuiltinJobs() {
	// Add(func(id ID) {
	// 	logger.Info("今天吃苹果了吗?")
	// }, "at 2025-12-25 20:08:00", "6 years", 1)

	// Add(DelayJobIfRunning(func(id ID) {
	// 	sender.SendToAdmin("登录支付宝领会员了", false)
	// }), "cron 0,15,30 59 9 1 * * *", "每月1号10点领会员", 0)

	Add(SkipJobIfRunning(func(id ID) {
		_ = rotatelog.Rotate()
	}), "cron 0 0 * * *", "每天零点切割日志", 0)

}
