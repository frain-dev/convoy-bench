# convoy-bench
convoy-bench contains all the code and scripts to benchmark any [convoy](https://github.com/frain-dev/convoy) cluster. The goal is to be able to quickly benchmark any Convoy cluster to know it's capacity. This was heavily inspired by [Clickbench](https://github.com/ClickHouse/ClickBench/) and [Lavinmqperf](https://lavinmq.com/documentation/lavinmqperf)

## Tools
1. k6.io 
2. Grafana
3. Ruby 3+
4. TimescaleDB

## Prequisites
1. Make sure you have a Convoy instance running and properly configured.
2. If you're going to be testing with a Message Broker, make sure you've provisioned the broker ahead as well. For each broker, there are specific parameters to set, read the flags to know what to supply.

## Usage
```bash
Usage:
  convoy-bench.rb exec -e, --endpoint-id=ENDPOINT-ID

Options:
  -p,   [--producer=PRODUCER]      # Select a producer to publish events from the following - http, sqs, pubsub or kafka.
                                   # Default: http
  -u,   [--uri=URI]                # URI of your Convoy Cluster.
                                   # Default: http://localhost:5005
  -v,   [--vus=VUS]                # Set how many virtual users should execute the test concurrently.
                                   # Default: 10
  -d,   [--duration=DURATION]      # Set how long the test should run. Use Golang string syntax: 1m, 5s, 10m5s .
                                   # Default: 1m
  -e,   --endpoint-id=ENDPOINT-ID  # ID of the endpoint configured on Convoy.
  -a,   [--api-key=API-KEY]        # Convoy Cluster API Key. Specify this if producer is http.
  -pid, [--project-id=PROJECT-ID]  # Convoy Cluster project ID. Specify this if producer is http.
  -q,   [--queue-url=QUEUE-URL]    # Amazon SQS URL. Specify this if producer is sqs.

execute convoy benchmarks
```

### Examples
The example below will run for 5 concurrent users for 5 minutes blasting events through Convoy's REST API.
```bash
./convoy-bench.rb exec -p http -v 5 -d 5m \
--endpoint-id "{endpoint-id}" \
--project-id "{project-id}"
--api-key "{api-key}" \
```

The example below will run for 5 concurrent users for 5 minutes blasting events through an Amazon SQS Queue.
```bash
» ./convoy-bench.rb exec -p sqs -v 5 -d 5m \
--endpoint-id "{endpoint-id}" \
--queue-url "{queue-url}" \
--aws-access-key "{aws-access-key}" \
--aws-secret-key "{aws-secret-key}"
```
