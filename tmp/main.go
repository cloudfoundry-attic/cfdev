package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/harlow/kinesis-consumer"
	"log"
	"time"
)

type StatMessage struct {
	Event string `json:"event"`
	Timestamp string `json:"timestamp"`
}

func main() {
	var stream = flag.String("cfdev-analytics-development", "cfdev-analytics-development", "cfdev-analytics-development")
	flag.Parse()

	// The following 2 lines will overwrite the default kinesis client
	myKinesisClient := kinesis.New(session.New(aws.NewConfig()), &aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(),
	})
	newKclient, err := consumer.NewKinesisClient(consumer.WithKinesis(myKinesisClient))

	// consumer
	c, err := consumer.New(
		*stream,
		consumer.WithClient(newKclient),
	)
	if err != nil {
		log.Fatalf("consumer error: %v", err)
	}
	ctx, _ := context.WithCancel(context.Background())
	fmt.Println(time.Now(), ": Starting.....")
	err = c.Scan(ctx, func(r *consumer.Record) consumer.ScanError {
		//fmt.Println(time.Now(), ":" ,string(r.Data))
		var dest StatMessage

		json.Unmarshal(r.Data, &dest)
		fmt.Printf("[%s]: %v\n", time.Now(), dest)
		eventTime, err := time.Parse(time.RFC3339, dest.Timestamp)
		tenMinutesAgo := time.Now().UTC().Add(-10 * time.Minute)
		fmt.Printf("current time: %v\n", time.Now().UTC())
		fmt.Printf("timestamp time: %v\n", eventTime)
		if eventTime.After(tenMinutesAgo) {
			fmt.Printf("found a new one")
		}
		err = errors.New("some error happened")
		// continue scanning
		return consumer.ScanError{
			Error:          err,
			StopScan:       false,
			SkipCheckpoint: false,
		}
	})
	if err != nil {
		fmt.Println("scan error: %v", err)
	}
}
