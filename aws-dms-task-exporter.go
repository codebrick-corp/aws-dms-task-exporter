package main

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	MetricName        = "task_stats"
	MetricNamespace   = "dms"
	MetricHelpMessage = "Gauge for dms tasks statistics"
	MetricLabels      = []string{"region", "identifier", "schema", "table", "action"}
)

type collector struct {
	sess       *session.Session
	dmsService *dms.DatabaseMigrationService
}

type DmsReplicationTask struct {
	arn        *string
	identifier *string
}

type DmsTaskStat struct {
	schemaName *string
	tableName  *string
	inserts    *int64
	deletes    *int64
	updates    *int64
}

func init() {
	viper.SetDefault("AWS_ACCESS_KEY_ID", "")     // must be set
	viper.SetDefault("AWS_SECRET_ACCESS_KEY", "") // must be set
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("AWS_REGION", "ap-southeast-1")
}

func main() {
	viper.AutomaticEnv()

	// Registers new collector to prometheus
	dmsCollector, err := NewCollector()
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(dmsCollector)

	// Creates new http handler
	http.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf("%s:%s", viper.Get("HOST"), viper.Get("PORT"))
	logrus.Info("Starting to listen ", addr)
	logrus.Fatal(http.ListenAndServe(addr, nil))
}

func NewCollector() (prometheus.Collector, error) {
	region := viper.GetString("AWS_REGION")
	sess, err := session.NewSession(&aws.Config{Region: &region})
	if err != nil {
		logrus.Fatal("Failed to create new aws session ", err.Error())
		return nil, err
	}
	dmsService := dms.New(sess)
	return &collector{sess, dmsService}, nil
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      MetricName,
		Namespace: MetricNamespace,
		Help:      MetricHelpMessage,
	}, MetricLabels)
	c.fetch(gauge)
	gauge.Collect(ch)
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc(
		prometheus.BuildFQName(MetricNamespace, "", MetricName),
		MetricHelpMessage,
		MetricLabels,
		nil,
	)
}

func (c *collector) fetch(gauge *prometheus.GaugeVec) {
	tasks, err := c.getDmsReplicationList()
	if err != nil {
		return
	}

	for _, task := range tasks {
		stats, err := c.getTableStatistics(task)
		if err != nil {
			continue
		}
		for _, stat := range stats {
			gauge.WithLabelValues(viper.GetString("AWS_REGION"), *task.identifier, *stat.schemaName, *stat.tableName, "insert").Set(float64(*stat.inserts))
			gauge.WithLabelValues(viper.GetString("AWS_REGION"), *task.identifier, *stat.schemaName, *stat.tableName, "delete").Set(float64(*stat.deletes))
			gauge.WithLabelValues(viper.GetString("AWS_REGION"), *task.identifier, *stat.schemaName, *stat.tableName, "update").Set(float64(*stat.updates))
		}
	}
}

func (c *collector) getDmsReplicationList() ([]DmsReplicationTask, error) {
	var result []DmsReplicationTask
	withoutSettings := true
	var marker *string = nil
	for {
		output, err := c.dmsService.DescribeReplicationTasks(&dms.DescribeReplicationTasksInput{
			Marker:          marker,
			WithoutSettings: &withoutSettings,
		})
		if err != nil {
			logrus.Fatal("Failed to get dms replication tasks, ", err.Error())
			return result, err
		}

		for _, task := range output.ReplicationTasks {
			result = append(result, DmsReplicationTask{
				arn:        task.ReplicationTaskArn,
				identifier: task.ReplicationTaskIdentifier,
			})
		}
		if output.Marker == nil {
			break
		} else {
			marker = output.Marker
		}
	}
	return result, nil
}

func (c *collector) getTableStatistics(task DmsReplicationTask) ([]DmsTaskStat, error) {
	var result []DmsTaskStat
	var marker *string = nil
	for {
		output, err := c.dmsService.DescribeTableStatistics(&dms.DescribeTableStatisticsInput{
			Marker:             marker,
			ReplicationTaskArn: task.arn,
		})
		if err != nil {
			logrus.Fatal("Failed to get table statistics, ", err.Error())
			return result, err
		}
		for _, stat := range output.TableStatistics {
			result = append(result, DmsTaskStat{
				schemaName: stat.SchemaName,
				tableName:  stat.TableName,
				inserts:    stat.Inserts,
				deletes:    stat.Deletes,
				updates:    stat.Updates,
			})
		}
		if output.Marker == nil {
			break
		} else {
			marker = output.Marker
		}
	}
	return result, nil
}
