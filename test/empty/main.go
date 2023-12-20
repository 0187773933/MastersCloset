package main

import (
	"fmt"
	"time"
	// "strings"
	// json "encoding/json"
	bolt_api "github.com/boltdb/bolt"
	// user "github.com/0187773933/MastersCloset/v1/user"
	// encryption "github.com/0187773933/MastersCloset/v1/encryption"
	uuid "github.com/satori/go.uuid"
)

func is_valid_uuid_v4( input string ) bool {
	// x_uuid , _ := uuid.FromBytes( input ) // never parses version ???
	x_uuid , _ := uuid.FromString( input )
	version := x_uuid.Version()
	return version == uuid.V4
}

// Delete Empty UUIDs ?
func main() {

	bolt_db_path := "/Users/morpheous/WORKSPACE/GO/MastersClosetTracker/mct.db"
	// bolt_db_key := ""
	db , _ := bolt_api.Open( bolt_db_path , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	db.Update( func( tx *bolt_api.Tx ) error {
		users_bucket := tx.Bucket( []byte( "users" ) )
		users_bucket.ForEach( func( x_uuid , bucket_value []byte ) error {
			if bucket_value == nil { return nil }
			s_x_uuid := string( x_uuid )
			if is_valid_uuid_v4( s_x_uuid ) != true {
				fmt.Println( "?? " , s_x_uuid )
				users_bucket.Delete( []byte( x_uuid ) )
			}
			return nil
		})
		return nil
	})
}
