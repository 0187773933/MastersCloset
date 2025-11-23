package server

import (
	"fmt"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	// bolt "github.com/0187773933/MastersCloset/v1/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	// pp "github.com/k0kubun/pp/v3"
	// pp.Println( viewed_user )
	// log "github.com/0187773933/MastersCloset/v1/log"
	logger "github.com/0187773933/MastersCloset/v1/logger"

	bolt_api "github.com/boltdb/bolt"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
)

// could move to user , but edit should be the only thing like this
func ( s *Server ) HandleUserEdit( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }
	var viewed_user user.User
	json.Unmarshal( context.Body() , &viewed_user )
	viewed_user.Config = &s.Config
	if viewed_user.DB == nil { viewed_user.DB = s.DB }
	viewed_user.Save()
	logger.Log.Info( fmt.Sprintf( "%s === Updated" , viewed_user.UUID ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/edit" ,
		"result": true ,
		"user": viewed_user ,
	})
}

func ( s *Server ) ImportUser( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }
	result := false
	uuid := context.Query( "uuid" )
	if uuid == "" { fmt.Println( "empty uuid" ); return context.Status( fiber.StatusBadRequest ).JSON( fiber.Map{ "result": result , } ) }
	body := context.Body()
	if len( body ) == 0 { fmt.Println( "empty body" ); return context.JSON( fiber.Map{ "result": result , } ) }
	db_result := s.DB.Update( func( tx *bolt_api.Tx ) error {
		users_bucket , users_bucket_err := tx.CreateBucketIfNotExists( []byte( "users" ) )
		if users_bucket_err != nil { fmt.Println( users_bucket_err ); return users_bucket_err }
		user_store_result := users_bucket.Put( []byte( uuid ) , body )
		if user_store_result != nil { fmt.Println( user_store_result ); return user_store_result }
		fmt.Println( "ImportUser - tracking change for user :" , uuid )
		return nil
	})
	if db_result != nil { fmt.Println( db_result ); return context.Status( 500 ).JSON( fiber.Map{ "result": result } ) }
	result = true
	return context.Status( 200 ).JSON( fiber.Map{
		"result": result ,
	})
}


func ( s *Server ) EditCheckIn( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }

	x_body := []byte( context.Body() )

	x_uuid := context.Params( "uuid" )
	x_ulid := context.Params( "ulid" )

	var x_checkin user.CheckIn
	json.Unmarshal( x_body , &x_checkin )
	fmt.Println( x_uuid , x_ulid , x_checkin )

	s.DB.Update( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( s.Config.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if len( viewed_user.CheckIns ) < 0 { return nil }
			for i , check_in := range viewed_user.CheckIns {
				if check_in.ULID == x_ulid {
					viewed_user.CheckIns[ i ] = x_checkin
					viewed_user_byte_object , _ := json.Marshal( viewed_user )
					viewed_user_byte_object_encrypted := encryption.ChaChaEncryptBytes( s.Config.BoltDBEncryptionKey , viewed_user_byte_object )
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