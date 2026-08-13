#!/bin/bash

# Default values
producer="http"
project="outgoing"
uri="http://localhost:5005"
vus="10"
rate="10"
duration="1m"
endpoint_id=""
api_key=""
project_id=""
queue_url=""
aws_access_key=""
aws_secret_key=""

# Function to display help
usage() {
  echo "Usage: $0 [options]"
  echo ""
  echo "Options:"
  echo "-p, --producer      Select a producer to publish events from - http, sqs, pubsub, kafka. Default: http"
  echo "-t, --project       Specify the project type, outgoing or incoming. Default: outgoing"
  echo "-u, --uri           Base URL for outgoing, ingest URL for incoming. Default: http://localhost:5005"
  echo "-v, --vus           Set how many virtual users should execute the test concurrently. Default: 10"
  echo "-r, --rate          Set how many requests should be sent per second. Default: 10"
  echo "-d, --duration      Set how long the test should run. Default: 1m"
  echo "-e, --endpoint-id   ID of the endpoint configured on Convoy."
  echo "-a, --api-key       Convoy Cluster API Key."
  echo "--project-id        Convoy Cluster project ID."
  echo "-q, --queue-url     Amazon SQS URL."
  echo "--aws-access-key    AWS Access Key."
  echo "--aws-secret-key    AWS Secret Key."
  echo "-h, --help          Display this help message."
}

# Parse command-line arguments
while [[ "$#" -gt 0 ]]; do
  case $1 in
    -p|--producer) producer="$2"; shift ;;
    -t|--project) project="$2"; shift ;;
    -u|--uri) uri="$2"; shift ;;
    -v|--vus) vus="$2"; shift ;;
    -r|--rate) rate="$2"; shift ;;
    -d|--duration) duration="$2"; shift ;;
    -e|--endpoint-id) endpoint_id="$2"; shift ;;
    -a|--api-key) api_key="$2"; shift ;;
    --project-id) project_id="$2"; shift ;;
    -q|--queue-url) queue_url="$2"; shift ;;
    --aws-access-key) aws_access_key="$2"; shift ;;
    --aws-secret-key) aws_secret_key="$2"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown parameter passed: $1"; usage; exit 1 ;;
  esac
  shift
done

# Set environment variables
export VUS="$vus"
export RATE="$rate"
export DURATION="$duration"
export BASE_URL="$uri"
export ENDPOINT_ID="$endpoint_id"
export PROJECT_ID="$project_id"
export API_KEY="$api_key"
export QUEUE_URL="$queue_url"
export AWS_ACCESS_KEY="$aws_access_key"
export AWS_SECRET_KEY="$aws_secret_key"

# Determine producer command based on producer type and project type
get_producer_command() {
  case $1 in
    http)
      if [[ "$project" == "outgoing" ]]; then
        echo "producer/http_test_outgoing.js"
      elif [[ "$project" == "incoming" ]]; then
        echo "producer/http_test_incoming.js"
      fi
      ;;
    sqs)
      echo "producer/sqs_test.js"
      ;;
    pubsub)
      echo "producer/pubsub_test.js"
      ;;
    kafka)
      echo "producer/kafka_test.js"
      ;;
    *)
      echo "Invalid producer type - $1" >&2
      exit 1
      ;;
  esac
}

producer_command=$(get_producer_command "$producer")

# bin/k6 is an x86_64 build, so on other architectures it runs under emulation,
# which distorts the timings the load generator itself reports. Prefer a k6
# installed for the host.
if command -v k6 >/dev/null 2>&1; then
  k6_bin="k6"
else
  k6_bin="./bin/k6"
  echo "warning: no k6 on PATH, falling back to the bundled x86_64 build in bin/k6." >&2
  echo "         install k6 for your architecture before recording results." >&2
fi

# Execute command
"$k6_bin" run "$producer_command"
