package utils

import "github.com/aws/aws-sdk-go-v2/service/sns/types"

func Ptr[V any](v V) *V { return &v }

const (
	DataTypeString string = "String"
	DataTypeNumber string = "Number"
)

func StringAttributeValue(s string) types.MessageAttributeValue {
	return types.MessageAttributeValue{
		DataType:    Ptr(DataTypeString),
		StringValue: Ptr(s),
	}
}

func NumberAttributeValue(s string) types.MessageAttributeValue {
	return types.MessageAttributeValue{
		DataType:    Ptr(DataTypeNumber),
		StringValue: Ptr(s),
	}
}
