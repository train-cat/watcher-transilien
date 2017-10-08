package metadata

import (
	"context"
	"os"

	"gopkg.in/olivere/elastic.v5"
	"github.com/spf13/viper"
	"github.com/train-sh/sniffer-transilien/utils"
)

type (
	indexable interface {
		getMappings() (string, string)
		Persist()
	}

	// station is struct for push to ElasticSearch
	station struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
		UIC  string `json:"uic"`
	}

	// Train is struct for push to ElasticSearch
	Train struct {
		Code    string `json:"code"`
		Mission string `json:"mission"`
	}
)

var (
	ctx = context.Background()
	client *elastic.Client

	indexes = []indexable{}
)

// Init client for elasticsearch
func Init() {
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(viper.GetString("elasticsearch.host")),
		elastic.SetSniff(false),
	}

	if viper.GetBool("elasticsearch.basic_auth") {
		options = append(options, elastic.SetBasicAuth(
			viper.GetString("elasticsearch.username"),
			viper.GetString("elasticsearch.password"),
		))
	}

	c, err := elastic.NewSimpleClient(options...)

	if err != nil {
		utils.Error(err.Error())
		os.Exit(utils.ErrorInitElasticSearch)
	}

	client = c

	// Create index and mappings
	for _, index := range indexes {
		idx, mapping := index.getMappings()

		exists, err := client.IndexExists(idx).Do(ctx)
		if err != nil {
			utils.Error(err.Error())
			os.Exit(utils.ErrorCreateIndexElasticSearch)
		}

		if !exists {
			_, err := client.CreateIndex(idx).BodyString(mapping).Do(ctx)
			if err != nil {
				utils.Error(err.Error())
				os.Exit(utils.ErrorCreateIndexElasticSearch)
			}
		}
	}
}
