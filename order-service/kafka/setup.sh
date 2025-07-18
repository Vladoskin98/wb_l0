#! /bin/bash
sleep 30

kafka-topics --bootstrap-server kafka:9092 --create --topic orders-topic --partitions 1 --replication-factor 1

echo "Kafka topic 'orders-topic' created"
