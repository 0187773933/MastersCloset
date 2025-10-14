package server

import (
	"fmt"
	"strings"
	"regexp"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	twilio "github.com/sfreiberg/gotwilio"
	log "github.com/0187773933/MastersCloset/v1/log"
	try "github.com/manucorporat/try"
)

func validate_us_phone_number( input string ) ( result string ) {
	if !strings.HasPrefix( input , "+1" ) { input = fmt.Sprintf( "+1%s" , input ) }
	input = strings.ReplaceAll( input , "-" , "" )
	r := regexp.MustCompile( "^\\+1[0-9]{10}$" )
	if !r.MatchString( input ) { result = "" } else { result = input }
	return result
}

func ( s *Server ) SMSAllUsers( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }
	// fmt.Println( context.GetReqHeaders() )
	sms_message := context.FormValue( "sms_message" )

	twilio_client := twilio.NewTwilioClient( s.Config.TwilioClientID , s.Config.TwilioAuthToken )

	s.DB.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( s.Config.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if viewed_user.PhoneNumber == "" { return nil; }

			validated_phone := validate_us_phone_number( viewed_user.PhoneNumber )
			if validated_phone == "" {
				log.Debug( fmt.Sprintf( "%s Has an Invalid phone number: %s" , viewed_user.NameString , viewed_user.PhoneNumber ) )
				return nil
			}
			// https://github.com/sfreiberg/gotwilio/blob/master/sms.go#L12
			try.This( func() {
				result , _ , _ := twilio_client.SendSMS( s.Config.TwilioSMSFromNumber , viewed_user.PhoneNumber , sms_message , "" , "" )
				log.Debug( fmt.Sprintf( "Texting === %s === %s\n" , validated_phone , result.Status ) )
			}).Catch( func( e try.E ) {
				log.Debug( fmt.Sprintf( "Failed to Text === %s === %s\n" , viewed_user.NameString , validated_phone ) )
			})

			return nil
		})
		return nil
	})

	return context.JSON( fiber.Map{
		"route": "/admin/user/sms/all" ,
		"sms_message": sms_message ,
		"result": "success" ,
	})
}


func ( s *Server ) SMSUser( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }
	// fmt.Println( context.GetReqHeaders() )
	sms_message := context.FormValue( "sms_message" )
	sms_number := context.FormValue( "sms_number" )
	validated_phone := validate_us_phone_number( sms_number )

	twilio_client := twilio.NewTwilioClient( s.Config.TwilioClientID , s.Config.TwilioAuthToken )

	if validated_phone == "" {
		log.Debug( fmt.Sprintf( "Invalid phone number: %s" , sms_number ) )
		return context.JSON( fiber.Map{
			"route": "/admin/user/sms" ,
			"sms_message": sms_message ,
			"to_number": sms_number ,
			"result": "invalid phone number" ,
		})
	}
	// https://github.com/sfreiberg/gotwilio/blob/master/sms.go#L12
	try.This( func() {
		result , _ , _ := twilio_client.SendSMS( s.Config.TwilioSMSFromNumber , sms_number , sms_message , "" , "" )
		log.Debug( fmt.Sprintf( "Texting === %s === %s\n" , validated_phone , result.Status ) )
	}).Catch(func( e try.E ) {
		log.Debug( fmt.Sprintf( "Failed to Text === %s\n" , validated_phone ) )
	})

	return context.JSON( fiber.Map{
		"route": "/admin/user/sms" ,
		"sms_message": sms_message ,
		"to_number": validated_phone ,
		"result": "success" ,
	})
}