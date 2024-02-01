package adminroutes

import (
	"fmt"
	"time"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	// bolt "github.com/0187773933/MastersCloset/v1/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	// pp "github.com/k0kubun/pp/v3"
	// pp.Println( viewed_user )
	// log "github.com/0187773933/MastersCloset/v1/log"

	bolt_api "github.com/boltdb/bolt"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
)

// could move to user , but edit should be the only thing like this
func HandleUserEdit( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	var viewed_user user.User
	json.Unmarshal( context.Body() , &viewed_user )
	viewed_user.Config = GlobalConfig
	viewed_user.Save();
	log.Info( fmt.Sprintf( "%s === Updated" , viewed_user.UUID ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/edit" ,
		"result": true ,
		"user": viewed_user ,
	})
}


func EditCheckIn( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	x_body := []byte( context.Body() )

	x_uuid := context.Params( "uuid" )
	x_ulid := context.Params( "ulid" )


	var x_checkin user.CheckIn
	json.Unmarshal( x_body , &x_checkin )
	fmt.Println( x_uuid , x_ulid , x_checkin )

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	db.Update( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if len( viewed_user.CheckIns ) < 0 { return nil }
			for i , check_in := range viewed_user.CheckIns {
				if check_in.ULID == x_ulid {
					viewed_user.CheckIns[ i ] = x_checkin
					viewed_user_byte_object , _ := json.Marshal( viewed_user )
					viewed_user_byte_object_encrypted := encryption.ChaChaEncryptBytes( GlobalConfig.BoltDBEncryptionKey , viewed_user_byte_object )
					bucket.Put( []byte( x_uuid ) , viewed_user_byte_object_encrypted )
					return nil
				}
			}
			return nil
		})
		return nil
	})

	return context.JSON( fiber.Map{
		"route": "/admin/checkins/edit/:uuid/:ulid" ,
		"uuid": x_uuid ,
		"ulid": x_ulid ,
		"result": true ,
	})
}