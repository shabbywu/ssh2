package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"ssh2/db"
)

type Model interface {
	GetId() int
	SetId(id int)
	GetName() string
	GetKind() string
	ToJson() ([]byte, error)
}

// jsonDumpAble: 将模型序列化到 json(数据库)
type jsonDumpAble struct {
	Kind string      `json:"kind,omitempty"yaml:"kind"`
	Spec interface{} `json:"spec,omitempty"yaml:"spec"`
}

func List[T interface{}](kind string) []T {
	objs := db.List(kind)
	var result []T

	for _, content := range objs {
		spec := gjson.Get(content, "spec").String()
		obj, err := parseObj[T](kind, spec)
		if err == nil {
			result = append(result, obj)
		} else {
			log.Fatal(err)
		}
	}
	return result
}

// Get Single Object By Field
func GetByField[T interface{}](kind string, field, value interface{}) (result T, err error) {
	content, err := db.GetByField(kind, field, value)
	if err != nil {
		return result, err
	}
	spec := gjson.Get(content, "spec").String()
	return parseObj[T](kind, spec)
}

func parseObj[T interface{}](kind, spec string) (result T, err error) {
	if e := json.Unmarshal([]byte(spec), &result); e != nil {
		return result, errors.New(fmt.Sprintf("非法的 %s 结构体", kind))
	}
	return result, nil
}

func GetByFieldGeneric(kind string, field, value interface{}) (result db.Model, err error) {
	switch kind {
	case "AuthMethod":
		obj, err := GetByField[AuthMethod](kind, field, value)
		return &obj, err
	case "ClientConfig":
		obj, err := GetByField[ClientConfig](kind, field, value)
		return &obj, err
	case "ServerConfig":
		obj, err := GetByField[ServerConfig](kind, field, value)
		return &obj, err
	case "Session":
		obj, err := GetByField[Session](kind, field, value)
		return &obj, err
	}
	return nil, fmt.Errorf("unknown kind: %s", kind)
}
