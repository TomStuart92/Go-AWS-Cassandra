package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/gocql/gocql"
)

var AWS AmazonWebServices
var cassandraClient *gocql.Session
var queueURL = "https://sqs.eu-west-1.amazonaws.com/355555488900/TomsTestQueue"

//AmazonWebServices provides a single point of contact with AWS SDK
type AmazonWebServices struct {
	Config   *aws.Config
	Session  *session.Session
	SQS      sqsiface.SQSAPI
	queueURL string
}

//IntializeCassandra creates session for reuse
func IntializeCassandra() *gocql.Session {
	fmt.Println("Intializing Cassandra")
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "tom"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	return session
}

//InitializeAWS create AWS Session
func InitializeAWS(queueURL string) AmazonWebServices {
	fmt.Println("Intializing AWS")
	config := &aws.Config{
		Region: aws.String("eu-west-1"),
	}
	session := session.Must(session.NewSession(config))
	SQS := sqs.New(session)
	return AmazonWebServices{config, session, SQS, queueURL}
}

//SendMessage sends a message to SQS
func (AWS *AmazonWebServices) SendMessage() {
	message := sqs.SendMessageInput{
		MessageBody: aws.String("Hello, World!"),
		QueueUrl:    &AWS.queueURL,
	}
	result, err := AWS.SQS.SendMessage(&message)
	if err != nil {
		fmt.Errorf("Error", err)
		return
	}
	fmt.Println("Successful Post. ID = ", *result.MessageId)
}

//ReadMessage reads a message from SQS
func (AWS *AmazonWebServices) ReadMessage(channel chan *sqs.Message) {
	message := sqs.ReceiveMessageInput{
		QueueUrl:            &AWS.queueURL,
		MaxNumberOfMessages: aws.Int64(1),
	}
	result, err := AWS.SQS.ReceiveMessage(&message)
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	if len(result.Messages) == 0 {
		fmt.Println("Received zero messages")
		return
	}
	fmt.Println("Successful GET.")
	channel <- result.Messages[0]
}

//PersistMessage takes a message via a channel and persists it to Cassandra
func PersistMessage(channel chan *sqs.Message) {
	for {
		message := <-channel
		messageID := *message.MessageId
		body := *message.Body

		// insert a record
		err := cassandraClient.Query(`INSERT INTO tom.test (id, body) VALUES (?, ?)`, messageID, body).Exec()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Saved 1 Record.")
	}
}

//LoopWithTimeout loops a function forever with a timeout between invocations
func LoopWithTimeout(functionToLoop func(), timeout int) {
	for {
		functionToLoop()
		time.Sleep(time.Second * time.Duration(timeout))
	}
}

//LoopWithChannel loops a function with a channel as an argument
func LoopWithChannel(functionToLoop func(chan *sqs.Message), channel chan *sqs.Message) {
	for {
		functionToLoop(channel)
	}
}

func main() {
	AWS = InitializeAWS(queueURL)
	cassandraClient = IntializeCassandra()
	defer cassandraClient.Close()

	go LoopWithTimeout(AWS.SendMessage, 5)

	messageChannel := make(chan *sqs.Message)
	go LoopWithChannel(AWS.ReadMessage, messageChannel)
	go LoopWithChannel(PersistMessage, messageChannel)

	var input string
	fmt.Scanln(&input)
}
