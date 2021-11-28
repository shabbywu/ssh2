package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"ssh2/db"
	"ssh2/models"
	"strconv"
)

//DocumentRecord: 序列化到文档的模型
type DocumentRecord struct {
	Kind string
	Spec map[interface{}]interface{}
	Ref  models.Ref
}

type YamlParser struct {
}

func parseIdField(spec map[interface{}]interface{}) (id int, err error) {
	if _id := spec["id"]; _id == nil {
		// ID 设置成 0, 存储时则自动生成 ID
		id = 0
	} else {
		switch v := _id.(type) {
		case int:
			id = v
		default:
			id, err = strconv.Atoi(v.(string))
			if err != nil {
				return 0, err
			}
		}
	}
	return id, err
}

func parserWrapper(core func(map[interface{}]interface{}) (models.Model, error)) func(map[interface{}]interface{}) (models.Model, error) {
	return func(m map[interface{}]interface{}) (models.Model, error) {
		parsed, err := core(m)
		if err != nil {
			return parsed, err
		}
		// TODO: save to db
		fmt.Printf("解析到 %s: %+v\n", parsed.GetKind(), parsed)
		_, err = db.UpdateOrCreate(parsed)
		return parsed, err
	}
}

func parseAuthMethodSpec(spec map[interface{}]interface{}) (models.Model, error) {
	var parsed = &models.AuthMethod{}

	id, err := parseIdField(spec)
	if err != nil {
		return parsed, err
	}

	parsed.ID = id
	parsed.Name = spec["name"].(string)
	parsed.Type = spec["type"].(string)
	parsed.Content = spec["content"].(string)

	if expect_for_password, ok := spec["expect_for_password"]; ok {
		parsed.ExpectForPassword = expect_for_password.(string)
	} else {
		parsed.ExpectForPassword = ""
	}
	return parsed, nil
}

func parseClientConfig(spec map[interface{}]interface{}) (models.Model, error) {
	var parsed = &models.ClientConfig{}

	id, err := parseIdField(spec)
	if err != nil {
		return parsed, err
	}

	parsed.ID = id
	parsed.Name = spec["name"].(string)
	parsed.User = spec["user"].(string)
	parsed.AuthMethodId = spec["auth_method_id"].(int)
	return parsed, nil
}

func parseServerConfig(spec map[interface{}]interface{}) (models.Model, error) {
	var parsed = &models.ServerConfig{}

	id, err := parseIdField(spec)
	if err != nil {
		return parsed, err
	}

	parsed.ID = id
	parsed.Name = spec["name"].(string)
	parsed.Host = spec["host"].(string)
	parsed.Port = spec["port"].(int)

	return parsed, nil
}

func parseSession(spec map[interface{}]interface{}) (models.Model, error) {
	var parsed = &models.Session{}

	id, err := parseIdField(spec)
	if err != nil {
		return parsed, err
	}

	var plugins []map[string]interface{}
	pluginsData := spec["plugins"].([]interface{})

	for _, data := range pluginsData {
		m := map[string]interface{}{}
		d := data.(map[interface{}]interface{})
		for k, v := range d {
			m[k.(string)] = v
		}
		plugins = append(plugins, m)
	}

	pluginsString, err := json.Marshal(plugins)
	if err != nil {
		return parsed, err
	}

	parsed.ID = id

	parsed.Name = spec["name"].(string)
	parsed.Tag = spec["tag"].(string)
	parsed.Plugins = string(pluginsString)

	parsed.ClientConfigId = spec["client_config_id"].(int)
	parsed.ServerConfigId = spec["server_config_id"].(int)
	return parsed, nil
}

var kindParserMapper = map[string]func(map[interface{}]interface{}) (models.Model, error){
	"AuthMethod":   parserWrapper(parseAuthMethodSpec),
	"ClientConfig": parserWrapper(parseClientConfig),
	"ServerConfig": parserWrapper(parseServerConfig),
	"Session":      parserWrapper(parseSession),
}

var attrKindMapper = map[string]string{
	"auth":   "AuthMethod",
	"client": "ClientConfig",
	"server": "ServerConfig",
}

var attrToFullKey = map[string]string{
	"auth":   "auth_method_id",
	"client": "client_config_id",
	"server": "server_config_id",
}

func (p YamlParser) ParseRecord(record DocumentRecord) (*models.Model, error) {
	spec := record.Spec
	for attr, kind := range attrKindMapper {
		if spec[attr] != nil {
			objDefinition := spec[attr].(map[interface{}]interface{})
			// 处理使用 ref 引用的逻辑
			if _ref := objDefinition["ref"]; _ref != nil {
				ref := _ref.(map[interface{}]interface{})
				obj, _ := models.GetByField(kind, ref["field"].(string), ref["value"])
				if instance, ok := obj.(models.Model); ok {
					// 替换属性成外键字段
					delete(spec, attr)
					spec[attrToFullKey[attr]] = instance.GetId()
				} else {
					log.Fatal(errors.New("404 Not Found"))
				}
			} else {
				// 递归解析属性
				innerSpec := objDefinition["spec"].(map[interface{}]interface{})

				innerRecord := DocumentRecord{
					Kind: kind,
					Spec: innerSpec,
				}

				instance, err := p.ParseRecord(innerRecord)
				if err != nil {
					log.Fatal(err)
				}
				// 替换属性成外键字段
				delete(spec, attr)
				spec[attrToFullKey[attr]] = (*instance).GetId()
			}
		}
	}

	if parser, ok := kindParserMapper[record.Kind]; ok {
		instance, err := parser(spec)
		if err != nil {
			log.Fatal(err)
		}
		return &instance, nil
	} else {
		return nil, errors.New("解析失败")
	}
}
