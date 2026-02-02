package mod

import (
	"bytes"
	"encoding/gob"
	"errors"
	"reflect"
	"unsafe"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type CustomFields map[string]any

const injectDataFieldID = 114514

func InjectDataToProto[T any](proto any, fieldName string, data T) {
	customFields := ExtractDataFromProto(proto)
	if customFields == nil {
		customFields = make(CustomFields)
	}
	customFields[fieldName] = data
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(customFields); err != nil {
		panic(err)
	}
	encodedData := buf.Bytes()
	unkFields, err := accessField[protoimpl.UnknownFields](proto, "unknownFields")
	if err != nil {
		panic(err)
	}

	*unkFields = protowire.AppendTag(nil, injectDataFieldID, protowire.BytesType)
	*unkFields = protowire.AppendBytes(*unkFields, encodedData)
}

func ExtractDataFromProto(proto any) (data CustomFields) {
	var customData CustomFields
	unkFields, err := accessField[protoimpl.UnknownFields](proto, "unknownFields")
	if err != nil {
		panic(err)
	}

	b := *unkFields
	if len(b) == 0 {
		return
	}
	num, typ, n := protowire.ConsumeTag(b)
	if n < 0 || num != injectDataFieldID || typ != protowire.BytesType {
		panic("Invalid tag in unknown fields")
	}
	v, _ := protowire.ConsumeBytes(b[n:])
	if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&customData); err != nil {
		panic(err)
	}
	return customData
}

func AccessDataFromProto[T any](proto any, fieldName string) (data T, found bool) {
	customFields := ExtractDataFromProto(proto)
	value, exists := customFields[fieldName]
	if !exists {
		return
	}
	return value.(T), true
}

func accessField[valueType any](obj any, fieldName string) (*valueType, error) {
	field := reflect.ValueOf(obj).Elem().FieldByName(fieldName)
	if field.Type() != reflect.TypeOf(*new(valueType)) {
		return nil, errors.New("field type: " + field.Type().String() + ", valueType: " + reflect.TypeOf(*new(valueType)).String())
	}
	v := (*valueType)(unsafe.Pointer(field.UnsafeAddr()))
	return v, nil
}
