package main

import (
	"github.com/segmentio/kafka-go"
	"testing"
)

func Test_ensureTopic(t *testing.T) {
	type args struct {
		broker string
		cfg    kafka.TopicConfig
	}
	tests := []struct {
		args args
	}{
		{
			args: args{
				broker: "localhost:9092",
				cfg: kafka.TopicConfig{
					Topic:             "order_id",
					NumPartitions:     3,
					ReplicationFactor: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		err := ensureTopic(tt.args.broker, tt.args.cfg)
		if err != nil {
			t.Fatal(err)
		}
	}
}
