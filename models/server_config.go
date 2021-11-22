package models

import "encoding/json"

type ServerConfig struct {
	ID int
	Name string
	Host string
	Port int
}

func (config *ServerConfig) ToJson () ([]byte, error){
	return json.Marshal(jsonDumpAble{
		Kind: "ServerConfig",
		Spec: config,
	})
}