package config

import "os"

type Config struct {
	AWSRegion      string
	DynamoEndpoint string
	RedisAddress   string
	RedisPass      string
}

func LoadConfig() *Config {
	return &Config{
		AWSRegion:      getEnv("AWS_REGION", "us-east-1"),
		DynamoEndpoint: getEnv("DYNAMO_ENDPOINT", "http://localhost:8000"),
		RedisAddress:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:      getEnv("REDIS_PASS", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
