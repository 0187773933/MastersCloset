package adminroutes

import (
	"fmt"
	// "strconv"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	// bolt "github.com/0187773933/MastersCloset/v1/bolt"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	// log "github.com/0187773933/MastersCloset/v1/log"
)

// http://localhost:5950/user/get/04b5fba6-6d76-42e0-a543-863c3f0c252c
func GetUser( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	user_uuid := context.Params( "uuid" )
	db := _get_db( context )
	viewed_user := user.GetByUUID( user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	log.Info( fmt.Sprintf( "%s === Selected" , viewed_user.UUID ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/:uuid" ,
		"result": viewed_user ,
	})
}

func GetUserViaBarcode( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	db := _get_db( context )
	barcode := context.Params( "barcode" )
	var viewed_user user.User
	db.View( func( tx *bolt_api.Tx ) error {
		barcode_bucket := tx.Bucket( []byte( "barcodes" ) )
		x_uuid := barcode_bucket.Get( []byte( barcode ) )
		if x_uuid == nil { return nil }
		log.Info( fmt.Sprintf( "Barcode : %s || UUID : %s" , barcode , x_uuid ) )
		user_bucket := tx.Bucket( []byte( "users" ) )
		x_user := user_bucket.Get( []byte( x_uuid ) )
		decrypted_user := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , x_user )
		json.Unmarshal( decrypted_user , &viewed_user )
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/barcode" ,
		"result": viewed_user ,
	})
}

func GetUserViaULID( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	db := _get_db( context )
	x_ulid := context.Params( "ulid" )
	var viewed_user user.User
	db.View( func( tx *bolt_api.Tx ) error {
		ulid_uuid_bucket := tx.Bucket( []byte( "ulid-uuid" ) )
		x_uuid := ulid_uuid_bucket.Get( []byte( x_ulid ) )
		if x_uuid == nil { return nil }
		log.Info( fmt.Sprintf( "ULID : %s || UUID : %s" , x_ulid , x_uuid ) )
		user_bucket := tx.Bucket( []byte( "users" ) )
		x_user := user_bucket.Get( []byte( x_uuid ) )
		decrypted_user := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , x_user )
		json.Unmarshal( decrypted_user , &viewed_user )
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/ulid" ,
		"result": viewed_user ,
	})
}

func GetAllUsers( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	// db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	// defer db.Close()
	db := _get_db( context )
	var result []user.GetUserResult
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			var get_user_result user.GetUserResult
			get_user_result.Username = viewed_user.Username
			get_user_result.UUID = viewed_user.UUID
			if len( viewed_user.CheckIns ) > 0 {
				get_user_result.LastCheckIn = viewed_user.CheckIns[ len( viewed_user.CheckIns ) - 1 ]
			}
			result = append( result , get_user_result )
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/all" ,
		"result": result ,
	})
}

func GetAllCheckIns( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	db := _get_db( context )
	date_totals := make(map[string]map[string]int)
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if len( viewed_user.CheckIns ) > 0 {
				for _, checkin := range viewed_user.CheckIns {
					if _, ok := date_totals[checkin.Date]; !ok {
						date_totals[checkin.Date] = make(map[string]int)
					}

					// Increment checkins count
					date_totals[checkin.Date]["checkins"]++

					// Increment shopped_for count
					if checkin.PrintJob.FamilySize > 0 {
						date_totals[checkin.Date]["shopped_for"] += checkin.PrintJob.FamilySize
					} else {
						date_totals[checkin.Date]["shopped_for"] += viewed_user.FamilySize
					}
					// fmt.Println( checkin.Date , date_totals[checkin.Date]["checkins"] , viewed_user.FamilySize , date_totals[checkin.Date]["shopped_for"] )
				}
			}
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/all/checkins" ,
		"result": date_totals ,
	})
}

func GetCheckinsDate( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	x_date := context.Params( "date" )

	db := _get_db( context )
	var result []user.CheckIn
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if len( viewed_user.CheckIns ) > 0 {
				for _ , check_in := range viewed_user.CheckIns {
					if check_in.Date == x_date {
						result = append( result , check_in )
					}
				}
			}
			return nil
		})
		return nil
	})

	return context.JSON( fiber.Map{
		"route": "/admin/checkins/get/:date" ,
		"date": x_date ,
		"result": result ,
	})
}

func GetCheckIn( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	x_uuid := context.Params( "uuid" )
	x_ulid := context.Params( "ulid" )

	// db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	// defer db.Close()
	db := _get_db( context )
	var result user.CheckIn
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if len( viewed_user.CheckIns ) < 0 { return nil }
			for _ , check_in := range viewed_user.CheckIns {
				if check_in.ULID == x_ulid {
					result = check_in
					return nil
				}
			}
			return nil
		})
		return nil
	})

	return context.JSON( fiber.Map{
		"route": "/admin/checkins/get/:uuid/:ulid" ,
		"uuid": x_uuid ,
		"ulid": x_ulid ,
		"result": result ,
	})
}


func GetAllEmails( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	// db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	// defer db.Close()
	db := _get_db( context )
	var result [][]string
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if viewed_user.EmailAddress == "" { return nil }
			x_user := []string{ viewed_user.UUID , viewed_user.NameString , viewed_user.EmailAddress }
			result = append( result , x_user )
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/all/emails" ,
		"result": result ,
	})
}

func GetAllPhoneNumbers( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	// db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	// defer db.Close()
	db := _get_db( context )
	var result [][]string
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if viewed_user.PhoneNumber == "" { return nil; }
			x_user := []string{ viewed_user.UUID , viewed_user.NameString , viewed_user.PhoneNumber }
			result = append( result , x_user )
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/all/phone-numbers" ,
		"result": result ,
	})
}

type UserBarcodeData struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Barcodes []string `json:"barcodes"`
}
func GetAllBarcodes( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	db := _get_db( context )
	var result []UserBarcodeData
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			if len( viewed_user.Barcodes ) < 1 { return nil }
			x_user := UserBarcodeData{
				UUID: viewed_user.UUID ,
				Name: viewed_user.NameString ,
				Barcodes: viewed_user.Barcodes ,
			}
			result = append( result , x_user )
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/get/all/barcodes" ,
		"result": result ,
	})
}