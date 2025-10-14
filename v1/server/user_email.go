package server

import (
	fmt "fmt"
	smtp "net/smtp"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	log "github.com/0187773933/MastersCloset/v1/log"
	try "github.com/manucorporat/try"
)

func ( s *Server ) SendEmail( to string , subject string , body string ) ( result bool ) {
	result = false
	try.This( func() {
		auth := smtp.PlainAuth( "" ,
			s.Config.Email.SMTPAuthEmail ,
			s.Config.Email.SMTPAuthPassword ,
			s.Config.Email.SMTPServer )
		msg := []byte( fmt.Sprintf( "From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s" , s.Config.Email.From , to , subject , body ) )
		fmt.Println( string( msg ) )
		err := smtp.SendMail( s.Config.Email.SMTPServerUrl , auth , s.Config.Email.From , []string{ to } , msg )
		if err != nil {
			fmt.Println( err )
		} else { result = true }
	}).Catch(func(e try.E) {
		// log.PrintfConsole( "Failed to Email === %s\n" , to )
		log.Info( fmt.Sprintf( "Failed to Email === %s\n" , to ) )
		fmt.Println( e )
	})
	return
}

func ( s *Server ) EmailUser( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }
	// fmt.Println( context.GetReqHeaders() )
	email_address := context.FormValue( "email-address" )
	email_subject := context.FormValue( "email-subject" )
	email_message := context.FormValue( "email-message" )

	email_result := s.SendEmail( email_address , email_subject , email_message )
	log.Info( fmt.Sprintf( "%s === %t" , email_address , email_result ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/email" ,
		"to": email_address ,
		"subject": email_subject ,
		"message": email_message ,
		"result": email_result ,
	})
}

func ( s *Server ) EmailAllUsers( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }
	// fmt.Println( context.GetReqHeaders() )

	email_subject := context.FormValue( "email-subject" )
	email_message := context.FormValue( "email-message" )

	result := true

	s.DB.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( s.Config.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if viewed_user.EmailAddress == "" { return nil; }
			// fmt.Println( viewed_user.EmailAddress , email_subject , email_message )
			email_result := s.SendEmail( viewed_user.EmailAddress , email_subject , email_message )
			if email_result == false { result = false }
			log.Info( fmt.Sprintf( "%s === %t" , viewed_user.EmailAddress , email_result ) )
			return nil
		})
		return nil
	})

	return context.JSON( fiber.Map{
		"route": "/admin/user/email/all" ,
		"subject": email_subject ,
		"message": email_message ,
		"result": result ,
	})
}