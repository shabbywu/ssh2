package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"ssh2/db"
	"ssh2/models"
	"strconv"
)

// DocumentRecord: 序列化到文档的模型
type DocumentRecord struct {
	Kind string                      `yaml:"kind"`
	Spec map[interface{}]interface{} `yaml:"spec"`
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
	parsed.Name = stringField(spec, "name")
	parsed.Type = stringField(spec, "type")
	parsed.Content = stringField(spec, "content")

	if expectForPassword, ok := spec["expect_for_password"]; ok {
		parsed.ExpectForPassword = fmt.Sprint(expectForPassword)
	} else {
		parsed.ExpectForPassword = ""
	}

	savePrivateKeyInDB := boolField(spec, "save_private_key_in_db")
	switch parsed.Type {
	case models.AuthPassword, models.AuthPublishKey, models.AUthPublishKeyFile, models.AUthInteractivePassword:
	default:
		return parsed, fmt.Errorf("unsupported auth type %s", parsed.Type)
	}
	if parsed.Type == models.AUthPublishKeyFile && savePrivateKeyInDB {
		content, err := os.ReadFile(parsed.Content)
		if err != nil {
			return parsed, err
		}
		parsed.Type = models.AuthPublishKey
		parsed.Content = string(content)
	}
	if parsed.Type == models.AUthInteractivePassword {
		parsed.Content = ""
	}
	if err := parsed.EncryptContent(); err != nil {
		return parsed, err
	}
	return parsed, nil
}

func stringField(spec map[interface{}]interface{}, name string) string {
	if value, ok := spec[name]; ok && value != nil {
		return fmt.Sprint(value)
	}
	return ""
}

func boolField(spec map[interface{}]interface{}, name string) bool {
	value, ok := spec[name]
	if !ok || value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		parsed, _ := strconv.ParseBool(v)
		return parsed
	default:
		return false
	}
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
	rawPlugins := spec["plugins"].([]interface{})

	for _, data := range rawPlugins {
		m := map[string]interface{}{}
		d := data.(map[interface{}]interface{})
		for k, v := range d {
			m[k.(string)] = v
		}
		plugins = append(plugins, m)
	}

	pluginsBytes, err := json.Marshal(plugins)
	if err != nil {
		return parsed, err
	}

	parsed.ID = id
	parsed.Plugins = string(pluginsBytes)
	parsed.Name = spec["name"].(string)
	parsed.Tag = spec["tag"].(string)
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
				obj, err := models.GetByFieldGeneric(kind, ref["field"].(string), ref["value"])
				if err != nil {
					return nil, fmt.Errorf("can't find model with kind=%s where %s=%s: %w", kind, ref["field"], ref["value"], err)
				}
				if instance, ok := obj.(models.Model); ok {
					// 替换属性成外键字段
					delete(spec, attr)
					spec[attrToFullKey[attr]] = instance.GetId()
				} else {
					return nil, fmt.Errorf("can't find model with kind=%s where %s=%s, detail: %s", kind, ref["field"], ref["value"], obj)
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
					return nil, err
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
			return nil, err
		}
		return &instance, nil
	} else {
		return nil, errors.New("解析失败")
	}
}
