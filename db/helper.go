package db

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/buntdb"
	"reflect"
)

// List All Object in db
func List(kind string) (result []string) {
	return ListByFilter(kind, func(key, value string) bool {
		return true
	})
}

// ListByFilter List All Object match the filter
func ListByFilter(kind string, filter func(key, value string) bool) (result []string) {
	index := fmt.Sprintf("%s:%s", kind, "id")
	db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend(index, func(key, value string) bool {
			if filter(key, value) {
				result = append(result, value)
			}
			return true
		})
	})
	return result
}

// Get Single Object By Field
func GetByField(kind, field, value interface{}) (result string, err error) {
	var filter string
	switch value.(type) {
	case int:
		filter = fmt.Sprintf(`{"spec": {"%s": %d}}`, field, value)
	default:
		filter = fmt.Sprintf(`{"spec": {"%s": "%s"}}`, field, value)
	}

	err = db.View(func(tx *buntdb.Tx) error {
		index := fmt.Sprintf("%s:%s", kind, field)
		return tx.AscendEqual(index, filter, func(key, value string) bool {
			result = value
			return false
		})
	})
	if result == "" {
		return "", fmt.Errorf("field %s=%s not exists", field, value)
	}
	return result, err
}

// GetMetaData 查询数据库中的元信息
func GetMetaData() (metadata MetaData, err error) {
	err = db.View(func(tx *buntdb.Tx) error {
		metadataString, err := tx.Get(MetaDataKey)
		err = json.Unmarshal([]byte(metadataString), &metadata)
		return err
	})
	return metadata, err
}

// GetNextId 返回下一个可用的 ID
func GetNextId(kind string) (result int, err error) {
	metadata, err := GetMetaData()
	if err != nil {
		return 1, err
	}
	field := reflect.ValueOf(metadata.ID).FieldByName(kind)
	result = int(field.Int()) + 1
	return result, nil
}
