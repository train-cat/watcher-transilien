package utils

const (
	_ = iota // 0 mean success
	ErrorInitElasticSearch
	ErrorCreateIndexElasticSearch
	ErrorInitDatabase
	ErrorInitQueue
	ErrorLoadTimezone
)
