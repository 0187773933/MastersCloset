package server

import (
	"fmt"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
)

func ( s *Server ) HandleUserSimilar( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }
	var sent_user user.User
	context_body := context.Body()
	// fmt.Println( string( context_body ) )
	json.Unmarshal( context_body , &sent_user )
	var similar_user_reports []user.UserSimilarReport
	s.DB.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( s.Config.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			similarity_report := sent_user.GetUserSimilarityReport( &viewed_user , s.Config.LevenshteinDistanceThreshold )
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

func ( s *Server ) HandleUserSimilarObjects( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }

	x_uuid := context.Params( "uuid" )
	viewed_user := user.GetByUUID( x_uuid , s.DB , s.Config.BoltDBEncryptionKey )
	viewed_user.GetSimilarUsers( &s.Config )
	fmt.Println( viewed_user )

	return context.JSON( fiber.Map{
		"route": "/admin/user/similar/o" ,
		"sent_user": viewed_user ,
	})
}