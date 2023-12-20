package adminroutes

import (
	fmt "fmt"
	time "time"
	smtp "net/smtp"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	log "github.com/0187773933/MastersCloset/v1/log"
	try "github.com/manucorporat/try"
)

func send_email( to string , subject string , body string ) ( result bool ) {
	result = false
	try.This( func() {
		auth := smtp.PlainAuth( "" ,
			GlobalConfig.Email.SMTPAuthEmail ,
			GlobalConfig.Email.SMTPAuthPassword ,
			GlobalConfig.Email.SMTPServer )
		msg := []byte( fmt.Sprintf( "From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s" , GlobalConfig.Email.From , to , subject , body ) )
		fmt.Println( string( msg ) )
		err := smtp.SendMail( GlobalConfig.Email.SMTPServerUrl , auth , GlobalConfig.Email.From , []string{ to } , msg )
		if err != nil {
			fmt.Println( err )
		} else { result = true }
	}).Catch(func(e try.E) {
		// log.PrintfConsole( "Failed to Email === %s\n" , to )
		fmt.Println( e )
	})
	return
}

func EmailUser( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	// fmt.Println( context.GetReqHeaders() )
	email_address := context.FormValue( "email-address" )
	email_subject := context.FormValue( "email-subject" )
	email_message := context.FormValue( "email-message" )

	email_result := send_email( email_address , email_subject , email_message )
	log.PrintlnConsole( email_address , fmt.Sprintf( "%t" , email_result ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/email" ,
		"to": email_address ,
		"subject": email_subject ,
		"message": email_message ,
		"result": email_result ,
	})
}

func EmailAllUsers( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	// fmt.Println( context.GetReqHeaders() )

	email_subject := context.FormValue( "email-subject" )
	email_message := context.FormValue( "email-message" )

	result := true
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if viewed_user.EmailAddress == "" { return nil; }
			// fmt.Println( viewed_user.EmailAddress , email_subject , email_message )
			email_result := send_email( viewed_user.EmailAddress , email_subject , email_message )
			if email_result == false { result = false }
			log.PrintlnConsole( viewed_user.EmailAddress , fmt.Sprintf( "%t" , email_result ) )
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