package main

import (
	"fmt"
	"time"
	"strings"
	json "encoding/json"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
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
			if strings.Contains( viewed_user.NameString , "Temp" ) {
				fmt.Println( "Deleting === " , viewed_user.NameString )
				users_bucket.Delete( []byte( x_uuid ) )
			}
			return nil
		})
		return nil
	})
}