# sniffer-transilien

Collect real time data from SNCF API (stop and train)

[![Go Report Card](https://goreportcard.com/badge/github.com/train-cat/sniffer-transilien)](https://goreportcard.com/report/github.com/train-cat/sniffer-transilien)

## Usage
### API SNCF
Get a api key [register](https://ressources.data.sncf.com/explore/dataset/api-temps-reel-transilien/)

```
cp config.json.dist config.json
# edit configuration 
go build -o sniffer-transilien *.go
./sniffer-transilien -config config.json
```

