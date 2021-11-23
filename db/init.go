package db

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/buntdb"
	"log"
)

var db *buntdb.DB

func init() {
	// Open the data.db file. It will be created if it doesn't exist.
	var err error
	db, err = buntdb.Open(":memory:")
	if err != nil {
		log.Fatal("Can't open db, detail: ", err)
	}
	indexes, err := db.Indexes()
	if err != nil {
		log.Fatal("Can't list indexes, detail: ", err)
	}

	if len(indexes) == 0 {
		initIndex()
	}
	initMetaData()
}

func initIndex() {
	db.Update(func(tx *buntdb.Tx) error {
		for _, index := range []struct {
			IndexName    string
			IndexPattern string
			IndexField   string
		}{
			{
				IndexName:    AuthMethodIdIndexName,
				IndexPattern: AuthMethodIndexPattern,
				IndexField:   "spec.id",
			},

			{
				IndexName:    AuthMethodNameIndexName,
				IndexPattern: AuthMethodIndexPattern,
				IndexField:   "spec.name",
			},

			{
				IndexName:    ClientConfigIdIndexName,
				IndexPattern: ClientConfigIndexPattern,
				IndexField:   "spec.id",
			},

			{
				IndexName:    ClientConfigNameIndexName,
				IndexPattern: ClientConfigIndexPattern,
				IndexField:   "spec.name",
			},

			{
				IndexName:    ServerConfigIdIndexName,
				IndexPattern: ServerConfigIndexPattern,
				IndexField:   "spec.id",
			},

			{
				IndexName:    ServerConfigNameIndexName,
				IndexPattern: ServerConfigIndexPattern,
				IndexField:   "spec.name",
			},

			{
				IndexName:    SessionIdIndexName,
				IndexPattern: SessionIndexPattern,
				IndexField:   "spec.id",
			},
			{
				IndexName:    SessionNameIndexName,
				IndexPattern: SessionIndexPattern,
				IndexField:   "spec.name",
			},
		} {
			if err := tx.CreateIndex(index.IndexName, index.IndexPattern, buntdb.IndexJSON(index.IndexField)); err != nil {
				log.Fatal(fmt.Sprintf("Fail to create index<`%s`> for pattern `%s`, detail: %s", index.IndexName, index.IndexPattern, err))
			}
		}
		return nil
	})
}

func initMetaData() {
	var err error
	err = db.Update(func(tx *buntdb.Tx) error {
		_, err = tx.Get(MetaDataKey)
		if err == buntdb.ErrNotFound {
			initData, err := json.Marshal(MetaData{})
			if err != nil {
				return err
			}
			_, _, err = tx.Set(MetaDataKey, string(initData), nil)
		}
		return nil
	})
	if err != nil {
		log.Fatal("Fail to init metadata, detail: ", err)
	}
}
