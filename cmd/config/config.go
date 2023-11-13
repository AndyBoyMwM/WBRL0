package config

import (
	"fmt"
	"os"
)

func SetenvConfig() {
	// NATS-Streaming Setenv
	err := os.Setenv("NATS_URL", "nats://127.0.0.1:4222")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("NATS_CLUSTER_ID", "test-cluster")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("NATS_CLIENT_ID", "client-1")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("NATS_SUBJECT", "channel-1")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("NATS_DURABLE_NAME", "durable-1")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("NATS_IDLE_TIMEOUT", "30")
	if err != nil {
		fmt.Println(err)
	}
	// DB Setenv
	err = os.Setenv("DB_HOST", "127.0.0.1:5432")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("DB_NAME", "WBR01")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("DB_PASS", "root")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("DB_USER", "root")
	if err != nil {
		fmt.Println(err)
	}
	// Cache Setenv
	err = os.Setenv("DB_CACHE_STREAM_SIZE", "5")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Setenv("DB_CACHE_STREAM", "stream-1")
	if err != nil {
		fmt.Println(err)
	}
}
