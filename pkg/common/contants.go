package common

// cronjob 历史保存条数

var K8SJobHistoryLimit int32 = 1

// cronjob 失败历史保存条数

var K8SJobFailedHistoryLimit int32 = 1

// 重试次数

var K8SJobBackoffLimit int32 = 1

// job 完成后保留时间。目前机制下，仅可为 0，即完成后立即删除。否则完成后立即再次执行，
// 会因名称重复而无法创建成功

var K8SJobTTLSeconds int32 = 0

// cronjob 完成后保留时间

var K8SCronJobTTLSeconds int32 = 30 * 60

// job 超时时间，必须大于 ttlSeconds

var K8SJobTimeoutSeconds int64 = 60 * 60

var K8SStartingDeadlineSeconds int64 = 60 * 5
