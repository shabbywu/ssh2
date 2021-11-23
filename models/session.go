package models

import (
	"encoding/json"
	"errors"
)

type Session struct {
	ID             int      `yaml:"id" json:"id,omitempty"`
	Name           string   `yaml:"name" json:"name,omitempty"`
	Tag            []string `yaml:"tag" json:"tag,omitempty"`
	ClientConfigId int      `yaml:"client_config_id" json:"client_config_id,omitempty"`
	ServerConfigId int      `yaml:"server_config_id" json:"server_config_id,omitempty"`

	Plugins string `yaml:"plugins" json:"plugins,omitempty"`
}

func (s Session) ToJson() ([]byte, error) {
	return json.Marshal(jsonDumpAble{
		Kind: s.GetKind(),
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

func (s Session) GetId() int {
	return s.ID
}

func (s Session) SetId(id int) {
	s.ID = id
}

func (s Session) GetName() string {
	return s.Name
}

func (s Session) GetKind() string {
	return "Session"
}
