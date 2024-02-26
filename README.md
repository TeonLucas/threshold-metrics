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
export METRIC_WHERE=NRQL_WHERE_CLAUSE
export METRIC_FACET=FACETS_TO_INCLUDE
export THRESHOLD=TARGET_THRESHOLD
export NEW_RELIC_LICENSE_KEY=YOUR_LICENSE_KEY
export NEW_RELIC_USER_KEY=YOUR_USER_API_KEY
```
You can optionally set the polling interval (default 1m):
```
export POLL_INTERVAL=5m
```

Let's look at a demo example NRQL query:
```
SELECT apm.mobile.ui.thread.duration, action, scope FROM Metric
WHERE appName = 'Acme Telco -Android' AND action LIKE '%onCreate'
```
The above query would be specified as follows:
```
export METRIC_NAME=apm.mobile.ui.thread.duration
export METRIC_WHERE="appName = 'Acme Telco -Android' AND action LIKE '%onCreate'"
export METRIC_FACET=action,scope
```
The `METRIC_NAME` specifies which field contains the timeslice aggregate values.
The `METRIC_WHERE` is applied as a NRQL WHERE clause to filter to specific entities and metrics.
The `METRIC_FACET` specifies which fields (`entity.guid` is automatically included) to apply as attributes on the dimensional metrics generated.

Then set a `THRESHOLD` x1, which will be used to find the area shown below:
![Threhold calculation](https://github.com/TeonLucas/threshold-metrics/blob/main/threshold-diagram.png)
For example,
```
export THRESHOLD=0.5
```

Then you can run as follows:
```
./threshold-metrics
```

This will generate a dimensional metric, calculating the count over threshold in a normal distribution, to each 1-minute timeslice metric aggregate.
