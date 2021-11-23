package db

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/buntdb"
	"reflect"
)

func UpdateOrCreate(model Model) (created bool, err error) {
	if model.GetId() == 0 {
		created = true
		id, err := getNextId(model.GetKind())
		if err != nil {
			return false, err
		}
		model.SetId(id)
	} else {
		created = false
	}

	dbKey := fmt.Sprintf("kind:%s:id:%d", model.GetKind(), model.GetId())
	value, err := model.ToJson()

	fmt.Printf("It's going to set %s to `%s`\n", value, dbKey)

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
	metadata, err := GetMetaData()
	if err != nil {
		return err
	}
	field := reflect.ValueOf(metadata.ID).FieldByName(model.GetKind())
	field.SetInt(int64(model.GetId()))

	metadataValue, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, _, err = tx.Set(MetaDataKey, string(metadataValue), nil)
	return err
}
