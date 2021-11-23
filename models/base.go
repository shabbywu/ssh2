package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"reflect"
	"ssh2/db"
)

type Model interface {
	GetId() int
	SetId(id int)
	GetName() string
	GetKind() string
	ToJson() ([]byte, error)
}

//jsonDumpAble: 将模型序列化到 json(数据库)
type jsonDumpAble struct {
	Kind string      `json:"kind,omitempty"yaml:"kind"`
	Spec interface{} `json:"spec,omitempty"yaml:"spec"`
}

type Ref struct {
	Field string
	Value interface{}
}

var kindTypeMap = map[string]reflect.Type{
	"AuthMethod":   reflect.TypeOf(AuthMethod{}),
	"ClientConfig": reflect.TypeOf(ClientConfig{}),
	"ServerConfig": reflect.TypeOf(ServerConfig{}),
	"Session":      reflect.TypeOf(Session{}),
}

func List(kind string) []interface{} {
	objs := db.List(kind)
	var result []interface{}

	for _, content := range objs {
		spec := gjson.Get(content, "spec").String()
		obj, err := parseObj(kind, spec)
		if err == nil {
			result = append(result, &obj)
		} else {
			log.Fatal(err)
		}
	}
	return result
}

// Get Single Object By Field
func GetByField(kind string, field, value interface{}) (result interface{}, err error) {
	content, err := db.GetByField(kind, field, value)
	if err != nil {
		return nil, err
	}
	spec := gjson.Get(content, "spec").String()
	return parseObj(kind, spec)
}

func parseObj(kind, spec string) (result interface{}, err error) {
	t := kindTypeMap[kind]

	instance := reflect.New(t)
	ptr := instance.Interface()

	if e := json.Unmarshal([]byte(spec), &ptr); e != nil {
		return nil, errors.New(fmt.Sprintf("非法的 %s 结构体", kind))
	}
	return ptr, nil
}
