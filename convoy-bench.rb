#!/usr/bin/env ruby

require 'rubygems'
require 'thor'

class ConvoyBench < Thor

  def self.exit_on_failure?
    true
  end

  desc "exec", "execute convoy benchmarks"
  option :producer, aliases: "-p", default: "http", 
    desc: "Select a producer to publish events from the following - http, sqs, pubsub or kafka."
  option :project, aliases: "-t", default: "outgoing",
      desc: "Specify the project type, outgoing or incoming"
  option :uri, aliases: "-u", default: "http://localhost:5005",
    desc: "For outgoing projects it is base url of your Convoy Cluster. For incoming projects it is the ingest url."
  option :vus, aliases: "-v", default: "10",
    desc: "Set how many virtual users should execute the test concurrently."
  option :rate, aliases: "-r", default: "10",
      desc: "Set how many requests should be sent per second."
  option :duration, aliases: "-d", default: "1m",
    desc: "Set how long the test should run. Use Golang string syntax: 1m, 5s, 10m5s."
  option "endpoint-id", aliases: "-e", required: false,
    desc: "ID of the endpoint configured on Convoy. Only specify this if the project is outgoing."
  option "api-key", aliases: "-a", required: false,
    desc: "Convoy Cluster API Key. Specify this if producer is http and the project is outgoing."
  option "project-id", aliases: "-pid", required: false,
    desc: "Convoy Cluster project ID. Specify this if producer is http and the project is outgoing."
  option "queue-url", aliases: "-q", required: false,
    desc: "Amazon SQS URL. Specify this if producer is sqs."
  option "aws-access-key", aliases: "-aak", required: false,
    desc: "AWS Access Key. Specify this if producer is sqs."
  option "aws-secret-key", aliases: "-ask", required: false,
    desc: "AWS Secret Key. Specify this if producer is sqs."
  def exec
    ENV["VUS"] = options[:vus]
    ENV["RATE"] = options[:rate]
    ENV["DURATION"] = options[:duration]
    ENV["BASE_URL"] = options[:uri]
    ENV["ENDPOINT_ID"] = options["endpoint-id"]
    ENV["PROJECT_ID"] = options["project-id"]
    ENV["API_KEY"] = options["api-key"]
    ENV["QUEUE_URL"] = options["queue-url"]
    ENV["AWS_ACCESS_KEY"] = options["aws-access-key"]
    ENV["AWS_SECRET_KEY"] = options["aws-secret-key"]

    exec_command = "./bin/k6"
    producer_command = get_producer_command(options[:producer])
    Kernel.system(exec_command, "run",  producer_command)
  end

  default_task :exec

  private

  def get_producer_command(producer)
    case producer
    when "http", :http
      if options[:project] == "outgoing"
        producer_command = "producer/http_test_outgoing.js"
      elsif options[:project] == "incoming"
        producer_command = "producer/http_test_incoming.js"
      end
    when "sqs", :sqs
      producer_command = "producer/sqs_test.js"
    when "pubsub", :pubsub 
      producer_command = "producer/pubsub_test.js"
    when "kafka", :kafka 
      producer_command = "producer/kafka_test.js"
    else
      raise ArgumentError.new "Invalid producer type - #{producer}"
    end

    producer_command
  end
end


ConvoyBench.start
