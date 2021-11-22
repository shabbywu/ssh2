package models

import (
	"encoding/json"
	"errors"
)

type Session struct {
	ID int
	Name string
	Tag []string
	ClientConfigId int
	ServerConfigId int

	Plugins string
}

func (s *Session) ToJson () ([]byte, error){
	return json.Marshal(jsonDumpAble{
		Kind: "Session",
		Spec: s,
	})
}


func (s *Session) ToExpectCommand() (cmd string, err error) {
	return "", nil
}

func (s *Session) GetClientConfig() (*ClientConfig, error) {
	spec := db.Where("kind", "=", "ClientConfig").Where("spec.id", "=", s.ClientConfigId).Select("spec").Get()
	obj, ok := spec.(ClientConfig)
	if !ok {
		return nil, errors.New("非法的 ClientConfig 结构体")
	}
	return &obj, nil
}

func (s *Session) GetServerConfig() (*ServerConfig, error) {
	spec := db.Where("kind", "=", "ServerConfig").Where("spec.id", "=", s.ServerConfigId).Select("spec").Get()
	obj, ok := spec.(ServerConfig)
	if !ok {
		return nil, errors.New("非法的 ServerConfig 结构体")
	}
	return &obj, nil
}
