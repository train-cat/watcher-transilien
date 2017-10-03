package utils

// List of exit code
const (
	_ = iota // 0 mean success
	ErrorInitElasticSearch
	ErrorCreateIndexElasticSearch
	ErrorInitDatabase
	ErrorInitQueue
	ErrorLoadTimezone
	ErrorInitPubSub
)
