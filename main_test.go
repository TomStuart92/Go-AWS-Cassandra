package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

type FakeSQS struct {
	sqsiface.SQSAPI
}

var hasBeenCalled bool

func (SQS *FakeSQS) SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	hasBeenCalled = true
	var strPointer = new(string)
	*strPointer = "ID1"
	message := sqs.SendMessageOutput{
		MessageId: strPointer,
	}
	return &message, nil
}
func TestSendMessagesShouldCall(t *testing.T) {
	testAWS := new(AmazonWebServices)
	testAWS.SQS = &FakeSQS{}
	testAWS.SendMessage()
	if !hasBeenCalled {
		t.Error("Send Message not Called")
	}
}
