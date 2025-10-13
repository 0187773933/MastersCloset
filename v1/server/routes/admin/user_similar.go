package adminroutes

import (
	"fmt"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
)

func HandleUserSimilar( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	var sent_user user.User
	context_body := context.Body()
	// fmt.Println( string( context_body ) )
	json.Unmarshal( context_body , &sent_user )
	var similar_user_reports []user.UserSimilarReport
	db := _get_db( context )
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			similarity_report := sent_user.GetUserSimilarityReport( &viewed_user , GlobalConfig.LevenshteinDistanceThreshold )
			if similarity_report.IsSimilar == false { return nil }
			similarity_report.User = viewed_user
			similar_user_reports = append( similar_user_reports , similarity_report )
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/similar" ,
		"sent_user": sent_user ,
		"similar_user_reports": similar_user_reports ,
	})
}

func HandleUserSimilarObjects( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	x_uuid := context.Params( "uuid" )
	db := _get_db( context )
	viewed_user := user.GetByUUID( x_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	viewed_user.GetSimilarUsers( GlobalConfig )
	fmt.Println( viewed_user )

	return context.JSON( fiber.Map{
		"route": "/admin/user/similar/o" ,
		"sent_user": viewed_user ,
	})
}