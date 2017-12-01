package utils

// List of exit code
const (
	_ = iota // 0 mean success
	ErrorInitElasticSearch
	ErrorCreateIndexElasticSearch
	ErrorInitQueue
	ErrorLoadTimezone
	ErrorInitPubSub
)
