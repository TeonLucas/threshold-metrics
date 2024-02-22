# threshold-metrics
Creates new metrics from NR timeslice metrics for outlier / threshold analysis.
Provide a timeslice metric name, and a threshold.  Dimensional metrics will be generated that
show how many data points are above the threshold.

## To build
```
go build
```

## To run
Set the following environment variables:
```
export NEW_RELIC_ACCOUNT=YOUR_ACCOUNT_ID
export METRIC_NAME=METRIC_TO_QUERY
export THRESHOLD=TARGET_THRESHOLD
export NEW_RELIC_LICENSE_KEY=YOUR_LICENSE_KEY
export NEW_RELIC_USER_KEY=YOUR_USER_API_KEY
```

Then you can run as follows:
```
./threshold-metrics
```
