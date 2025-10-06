package db

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
	"reflect"
)

func UpdateOrCreate(model Model) (created bool, err error) {
	if model.GetId() == 0 {
		sameNameModel, _ := GetByField(model.GetKind(), "name", model.GetName())
		if sameNameModel != "" {
			created = false
			id := gjson.Get(sameNameModel, "spec.id").Int()
			model.SetId(int(id))
		} else {
			created = true
			id, err := GetNextId(model.GetKind())
			if err != nil {
				return false, err
			}
			model.SetId(id)
		}
	} else {
		created = false
	}

	dbKey := fmt.Sprintf("kind:%s:id:%d", model.GetKind(), model.GetId())
	value, err := model.ToJson()

	if err != nil {
		return false, err
	}

	err = db.Update(func(tx *buntdb.Tx) error {
		_, _, err = tx.Set(dbKey, string(value), nil)
		if err != nil {

			return err
		}

		if created {
			return updateMetaID(model, tx)
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	return created, nil
}

func updateMetaID(model Model, tx *buntdb.Tx) error {
	metadata := &MetaData{}
	metadataString, err := tx.Get(MetaDataKey)
	err = json.Unmarshal([]byte(metadataString), metadata)

	if err != nil {
		return err
	}
	field := reflect.ValueOf(&metadata.ID).Elem().FieldByName(model.GetKind())
	field.SetInt(int64(model.GetId()))

	metadataValue, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, _, err = tx.Set(MetaDataKey, string(metadataValue), nil)
	return err
}
