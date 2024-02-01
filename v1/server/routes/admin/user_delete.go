package adminroutes

import (
	"fmt"
	"time"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	bleve "github.com/blevesearch/bleve/v2"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	// log "github.com/0187773933/MastersCloset/v1/log"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
)

func DeleteUser( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	user_uuid := context.Params( "uuid" )

	search_index , _ := bleve.Open( GlobalConfig.BleveSearchPath )
	defer search_index.Close()
	search_index.Delete( user_uuid )

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	viewed_user := user.GetByUUID( user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	db.Update( func( tx *bolt_api.Tx ) error {
		users_bucket := tx.Bucket( []byte( "users" ) )
		users_bucket.Delete( []byte( user_uuid ) )
		usernames_bucket := tx.Bucket( []byte( "usernames" ) )
		usernames_bucket.Delete( []byte( viewed_user.Username ) )
		return nil
	})
	log.Info( fmt.Sprintf( "%s === Deleted" , viewed_user.UUID ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/delete/:uuid" ,
		"result": "deleted" ,
	})
}

func DeleteCheckIn( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	user_uuid := context.Params( "uuid" )
	check_in_ulid := context.Params( "ulid" )
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	// viewed_user := user.GetByUUID( user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	db.Update( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		var viewed_user user.User
		decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		if len( viewed_user.CheckIns ) < 1 { fmt.Println( "???" ); return nil }
		for i , check_in := range viewed_user.CheckIns {
			if check_in.ULID == check_in_ulid {
				viewed_user.CheckIns = append( viewed_user.CheckIns[ :i ] , viewed_user.CheckIns[ i+1 : ]... )
				log.Info( fmt.Sprintf( "%s === %s === Deleted" , viewed_user.UUID , check_in_ulid ) )
				break;
			}
		}
		viewed_user_byte_object , _ := json.Marshal( viewed_user )
		viewed_user_byte_object_encrypted := encryption.ChaChaEncryptBytes( GlobalConfig.BoltDBEncryptionKey , viewed_user_byte_object )
		bucket.Put( []byte( user_uuid ) , viewed_user_byte_object_encrypted )

		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/checkins/delete/:uuid/:ulid" ,
		"result": "deleted" ,
	})
}