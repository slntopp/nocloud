package eventbus

import (
	"fmt"

	pb "github.com/slntopp/nocloud-proto/events"
)

const (
	TOPIC_FORMAT = "%s.%s"
)

func Topic(msg interface{}) string {

	switch v := msg.(type) {
	case *pb.ConsumeRequest:
		return fmt.Sprintf(TOPIC_FORMAT, v.Type, v.Uuid)
	case *pb.Event:
		return fmt.Sprintf(TOPIC_FORMAT, v.Type, v.Uuid)
	case *pb.CancelRequest:
		return fmt.Sprintf(TOPIC_FORMAT, v.Type, v.Uuid)
	default:
		return ""
	}
}
