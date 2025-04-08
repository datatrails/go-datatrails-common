package azbus

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	otrace "github.com/opentracing/opentracing-go"
)

type MsgSender interface {
	Open() error
	Close(context.Context)

	Send(context.Context, *OutMessage) error
	NewMessageBatch(context.Context) (*OutMessageBatch, error)
	BatchAddMessage(context.Context, otrace.Span, *OutMessageBatch, *OutMessage, *azservicebus.AddMessageOptions) error

	SendBatch(context.Context, *OutMessageBatch) error
	String() string
}
