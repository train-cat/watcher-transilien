# train-sniffer

Collect data from SNCF API (passage and train) and real time information. Keep all information in MySQL and ElasticSearch

## Usage
### API SNCF
Get a api key [register](https://ressources.data.sncf.com/explore/dataset/api-temps-reel-transilien/)

```
cp config.json.dist config.json
# edit configuration 
go build -o train-sniffer *.go
./train-sniffer -config config.json
```

