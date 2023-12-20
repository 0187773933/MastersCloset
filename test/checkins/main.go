package main

import (
	// "fmt"
	"time"
	// "strings"
	json "encoding/json"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	ulid "github.com/oklog/ulid/v2"
)

func main() {

	bolt_db_path := "/Users/morpheous/WORKSPACE/GO/MastersClosetTrackerDecentralized/mct.db"
	bolt_db_key := ""
	db , _ := bolt_api.Open( bolt_db_path , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	db.Update( func( tx *bolt_api.Tx ) error {
		users_bucket := tx.Bucket( []byte( "users" ) )
		users_bucket.ForEach( func( x_uuid , bucket_value []byte ) error {
			if bucket_value == nil { return nil }
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( bolt_db_key , bucket_value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )

			if len( viewed_user.CheckIns ) < 1 { return nil }
			for i := 0 ; i < len( viewed_user.CheckIns ); i++ {
				// viewed_user.CheckIns[i].UUID = viewed_user.UUID
				// viewed_user.CheckIns[i].Name = viewed_user.NameString
				if viewed_user.CheckIns[i].ULID == "" {
					viewed_user.CheckIns[i].ULID = ulid.Make().String()
				}
				if viewed_user.CheckIns[i].PrintJob.FamilySize == 0 {
					viewed_user.CheckIns[i].PrintJob.FamilySize = viewed_user.FamilySize
				}
			}

			viewed_user_byte_object , _ := json.Marshal( viewed_user )
			viewed_user_byte_object_encrypted := encryption.ChaChaEncryptBytes( bolt_db_key , viewed_user_byte_object )
			users_bucket.Put( []byte( x_uuid ) , viewed_user_byte_object_encrypted )
			return nil
		})
		return nil
	})
}
