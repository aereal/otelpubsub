package pub

import (
	"maps"
	"slices"
	"sync"

	"github.com/aereal/otelpubsub/amazonsns/internal/utils"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"go.opentelemetry.io/otel/propagation"
)

// NewMessageAttributeCarrier wraps SNS message attributes as a [propagation.TextMapCarrier].
// Only String-typed attributes are accessible for reading; other types are ignored.
func NewMessageAttributeCarrier(attrs map[string]types.MessageAttributeValue) propagation.TextMapCarrier {
	return &carrier{
		attributes: attrs,
		strs: sync.OnceValue(func() map[string]string {
			ret := map[string]string{}
			for k, mav := range attrs {
				if mav.DataType == nil || *mav.DataType != utils.DataTypeString {
					continue
				}
				if mav.StringValue == nil {
					continue
				}
				sv := *mav.StringValue
				ret[k] = sv
			}
			return ret
		}),
	}
}

type carrier struct {
	attributes map[string]types.MessageAttributeValue
	strs       func() map[string]string
}

var _ propagation.TextMapCarrier = (*carrier)(nil)

func (c *carrier) Get(key string) string {
	return c.strs()[key]
}

func (c *carrier) Set(key, value string) {
	c.attributes[key] = utils.StringAttributeValue(value)
}

func (c *carrier) Keys() []string {
	return slices.Sorted(maps.Keys(c.strs()))
}
