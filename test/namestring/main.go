package main

import (
	"fmt"
	"time"
	// "strings"
	json "encoding/json"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	bleve "github.com/blevesearch/bleve/v2"
	types "github.com/0187773933/MastersCloset/v1/types"
)

func main() {

	bolt_db_path := "/Users/morpheous/WORKSPACE/GO/MastersClosetTrackerDecentralized/mct.db"
	bolt_db_key := ""
	db , _ := bolt_api.Open( bolt_db_path , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	search_index , _ := bleve.Open( "/Users/morpheous/WORKSPACE/GO/MastersClosetTrackerDecentralized/mct.bleve" )
	defer search_index.Close()
	db.Update( func( tx *bolt_api.Tx ) error {
		users_bucket := tx.Bucket( []byte( "users" ) )
		stats := users_bucket.Stats()
		index := 1
		users_bucket.ForEach( func( x_uuid , bucket_value []byte ) error {
			fmt.Printf( "Checking User [ %d ] of %d\n" , index , stats.KeyN )
			index += 1
			if bucket_value == nil { return nil }
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( bolt_db_key , bucket_value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )

			original_username := viewed_user.Username
			original_name_string := viewed_user.NameString

			viewed_user.FormatUsername()
			fmt.Println( "\t" , original_username , "===" , viewed_user.Username )
			fmt.Println( "\t" , original_name_string , "===" , viewed_user.NameString )

			if original_name_string != viewed_user.NameString {
				fmt.Println( "\tUpdating NameString" , original_name_string , "to" , viewed_user.NameString )
			}

			usernames_bucket , _ := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
			if original_username != viewed_user.Username {
				fmt.Println( "\tUpdating Username" , original_username , "to" , viewed_user.Username )
				usernames_bucket.Delete( []byte( original_username ) )
				edited_search_item := types.SearchItem{
					UUID: viewed_user.UUID ,
					Name: viewed_user.NameString ,
				}
				search_index.Index( viewed_user.UUID , edited_search_item )
			}
			usernames_bucket.Put( []byte( viewed_user.Username ) , []byte( viewed_user.UUID ) )

			byte_object , _ := json.Marshal( viewed_user )
			byte_object_encrypted := encryption.ChaChaEncryptBytes( bolt_db_key , byte_object )
			users_bucket.Put( []byte( viewed_user.UUID ) , byte_object_encrypted )

			return nil
		})
		return nil
	})
}
