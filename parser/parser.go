package parser

import (
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

func parseOrGenerateId(kind string, spec map[interface{}]interface{}) (id int, err error) {
	if _id := spec["id"]; _id == nil {
		// TODO: 生成 id
		id = 1
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
	var parsed = models.AuthMethod{}

	id, err := parseOrGenerateId("AuthMethod", spec)
	if err != nil {
		return parsed, err
	}

	parsed.ID = id
	parsed.Name = spec["name"].(string)
	parsed.Type = spec["type"].(string)
	parsed.Content = spec["content"].(string)
	parsed.ExpectForPassword = spec["expect_for_password"].(string)
	return parsed, nil
}

func parseClientConfig(spec map[interface{}]interface{}) (models.Model, error) {
	var parsed = models.ClientConfig{}

	id, err := parseOrGenerateId("ClientConfig", spec)
	if err != nil {
		return parsed, err
	}

	parsed.ID = id
	parsed.Name = spec["name"].(string)
	parsed.User = spec["user"].(string)
	parsed.AuthMethodId = spec["auth_method_id"].(int)
	return parsed, nil
}

var kindParserMapper = map[string]func(map[interface{}]interface{}) (models.Model, error){
	"AuthMethod":   parserWrapper(parseAuthMethodSpec),
	"ClientConfig": parserWrapper(parseClientConfig),
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
			if _ref := objDefinition["ref"]; _ref != nil {
				ref := _ref.(map[interface{}]interface{})
				obj, _ := models.GetByField(kind, ref["field"].(string), ref["value"])
				if instance, ok := obj.(models.Model); !ok {
					delete(spec, attr)
					spec[attrToFullKey[attr]] = instance.GetId()
				} else {
					log.Fatal(errors.New("404 Not Found"))
				}
			} else {
				innerSpec := objDefinition["spec"].(map[interface{}]interface{})
				if parser, ok := kindParserMapper[kind]; ok {
					instance, err := parser(innerSpec)
					if err != nil {
						log.Fatal(err)
					}
					delete(spec, attr)
					spec[attrToFullKey[attr]] = instance.GetId()
				}
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
