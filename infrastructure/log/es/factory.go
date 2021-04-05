package es

import (
	"context"
	"flag"
	"fmt"
	"github.com/huskar-t/gopher/infrastructure/log"
	"github.com/huskar-t/gopher/infrastructure/log/query"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"strings"
	"time"
)

type LoggerFactory struct {
	client    *elastic.Client
	formatter logrus.Formatter
	level     logrus.Level
	index     string
	host      string
}

func (factory *LoggerFactory) CreateHook() (logrus.Hook, error) {
	return NewBulkProcessorElasticHook(factory.client, factory.host, logrus.InfoLevel, factory.index)
}

func (factory *LoggerFactory) SetHost(host string) {
	factory.host = host
}

func (factory *LoggerFactory) SetFormatter(formatter logrus.Formatter) {
	factory.formatter = formatter
}

func (factory *LoggerFactory) SetLevel(level logrus.Level) {
	factory.level = level
}

func (factory *LoggerFactory) Query(index, host, module, level, content string, from, to time.Time, offset, limit int, tagCond *query.Condition) (total int64, items []log.Message, err error) {
	q := elastic.NewBoolQuery()
	if host != "" {
		q.Must(elastic.NewTermQuery("host", host))
	}
	if module != "" {
		q.Must(elastic.NewTermQuery("module", module))
	}

	if level != "" {
		q.Must(elastic.NewTermsQuery("level", level))
	}

	if tagCond != nil {
		if tagCond.Ands != nil {
			for k, v := range tagCond.Ands {
				if k == "log_username" {
					q.Must(elastic.NewPrefixQuery("tags.log_username", strings.ToLower(v.(string))))
				} else {
					q.Must(elastic.NewMatchQuery(fmt.Sprintf("tags.%s", k), v))
				}
			}
		}

		if len(tagCond.Ors) > 0 {
			var should []elastic.Query
			var orQuery = elastic.NewBoolQuery()
			for _, fields := range tagCond.Ors {
				subAndQuery := elastic.NewBoolQuery()
				if len(fields) > 0 {
					for k, v := range fields {
						if k == "log_username" {
							subAndQuery.Must(elastic.NewPrefixQuery("tags.log_username", strings.ToLower(v.(string))))
						} else {
							subAndQuery.Must(elastic.NewMatchQuery(fmt.Sprintf("tags.%s", k), v))
						}
					}
					should = append(should, subAndQuery)
				}
			}
			if len(should) > 0 {
				orQuery.Should(should...)
				q.Must(orQuery)
			}
		}
	}
	if content != "" {
		q.Must(elastic.NewMatchQuery("message", content))
	}
	q.Must(elastic.NewRangeQuery("timestamp").TimeZone("UTC").From(from.UTC().Format(time.RFC3339Nano)).To(to.UTC().Format(time.RFC3339Nano)))

	search := factory.client.Search().
		Index(index).
		Sort("timestamp", false).
		Query(q)
	if limit > 0 {
		search = search.Size(limit)
	}
	if offset > 0 {
		search = search.From(offset)
	}
	searchResult, err := search.Do(context.Background())
	if err != nil {
		return 0, nil, err
	}

	for _, item := range searchResult.Each(reflect.TypeOf(log.Message{})) {
		m := item.(log.Message)
		items = append(items, m)
	}
	total = searchResult.TotalHits()
	return
}

var addr = "http://localhost:9200"

func CreateFactory(app string) *LoggerFactory {
	if addr == "" {
		logrus.Fatal("elastic address is empty")
	}
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(strings.Split(addr, ",")...))
	if err != nil {
		logrus.Fatalf("connect elastic error: %+v", err)
	}
	logrus.RegisterExitHandler(func() {
		client.Stop()
	})
	return &LoggerFactory{
		client:    client,
		formatter: &logrus.TextFormatter{},
		level:     logrus.InfoLevel,
		index:     app,
	}
}

func init() {
	if s := os.Getenv("ES_ADDR"); s != "" {
		addr = s
	}
	flag.StringVar(&addr, "es.addr", addr, "elasticsearch listen address")
}
