package main

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/stretchr/testify/mock"
)

type FakeSQS struct {
	mock.Mock
	sqsiface.SQSAPI
}

func (SQS *FakeSQS) SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	args := SQS.Called(input)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func (SQS *FakeSQS) ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	args := SQS.Called(input)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func TestSendMessagesShouldCallSQSSendMessage(t *testing.T) {
	testAWS := new(AmazonWebServices)
	fakeSQS := FakeSQS{}

	response := new(sqs.SendMessageOutput)
	fakeSQS.On("SendMessage", mock.Anything).Return(response, nil)

	testAWS.SQS = &fakeSQS
	testAWS.SendMessage()

	fakeSQS.AssertExpectations(t)
}

func TestSendMessagesShouldNotThrowOnError(t *testing.T) {
	testAWS := new(AmazonWebServices)
	fakeSQS := FakeSQS{}

	sampleErr := errors.New("Something Went Wrong With AWS")
	response := new(sqs.SendMessageOutput)
	fakeSQS.On("SendMessage", mock.Anything).Return(response, sampleErr)

	testAWS.SQS = &fakeSQS
	testAWS.SendMessage()

	fakeSQS.AssertExpectations(t)
}
