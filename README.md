# aws-dms-task-exporter

 This is simple service to scrapes the [AWS DMS Task](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Monitoring.html), especially for [DMS Table Statistics](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Monitoring.html#CHAP_Tasks.CustomizingTasks.TableState). The exporter exports the prometheus metrics via HTTP. It could help you to monitor detailed metrics about AWS DMS tasks.


## Installation

### Docker
```shell
docker run -d \
 -p 8080:8080 \
 -e AWS_ACCESS_KEY_ID='YOUR AWS KEY ID' \
 -e AWS_SECRET_ACCESS_KEY='YOUR AWS SECRET ACCESS KEY' \
 ghcr.io/codebrick-corp/aws-dms-task-exporter:main
```

### Local
```shell
# Build aws-dms-task-exporter
go mod tidy && go build .

# Set environments for AWS credentials
expose AWS_ACCESS_KEY_ID='YOUR AWS KEY ID'
expose AWS_SECRET_ACCESS_KEY='YOUR AWS SECRET ACCESS KEY'

# Run an exporter
./aws-dms-task-exporter 
```

## Metrics
Belows are the list of metrics that `aws-dms-task-exporter` exports.

Sample metrics
```
# TYPE dms_task_stats gauge
dms_task_stats{action="delete",region="ap-southeast-1",schema="example_schema",table="inventories"} 40601
dms_task_stats{action="insert",region="ap-southeast-1",schema="example_schema",table="inventories"} 4.145428e+06
dms_task_stats{action="update",region="ap-southeast-1",schema="example_schema",table="inventories"} 1.24051e+06
```

Name | Description | Labels
-----|-----|-----
dms_task_stats | DMS Task Table Statistics showing counts of Insert, Delete, Update of source tables | action, region, schema, table
