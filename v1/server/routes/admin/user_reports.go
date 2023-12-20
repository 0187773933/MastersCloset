package adminroutes

import (
	// "fmt"
	"time"
	"bytes"
	"strconv"
	// "reflect"
	csv "encoding/csv"
	json "encoding/json"
	fiber "github.com/gofiber/fiber/v2"
	// bolt "github.com/0187773933/MastersCloset/v1/bolt"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	log "github.com/0187773933/MastersCloset/v1/log"
)

// https://mailchimp.com/help/import-contacts-mailchimp/
// https://mailchimp.com/help/format-guidelines-for-your-import-file/#Email_address
func GetReportMailChimp( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	// Get Data from DB
	var result [][]string
	result = append( result , []string{ "First Name" , "Last Name" , "Email Address" , "Phone" } )
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			x_user := []string{ viewed_user.Identity.FirstName , viewed_user.Identity.LastName , viewed_user.EmailAddress , viewed_user.PhoneNumber }
			result = append( result , x_user )
			return nil
		})
		return nil
	})

	// Build CSV File
	var csv_buffer bytes.Buffer
	writer := csv.NewWriter( &csv_buffer )
	for _ , row := range result {
		writer.Write( row )
	}
	writer.Flush()
	csv_bytes := csv_buffer.Bytes()
	context.Set( "Content-Type", "text/csv" )
	context.Set( "Content-Disposition", "attachment;filename=masters_closet_contacts_mail_chimp.csv" )
	context.Set( "Content-Length", strconv.Itoa( len( csv_bytes ) ) )
	log.PrintlnConsole( "Downloaded Report-MailChimp" , strconv.Itoa( len( csv_bytes ) ) )
	return context.Send( csv_bytes )

}

// This would automatically flatten all struct fields , but were just going to manually do it ....
// csv_headers := []string{}
// extract_fields( "" , reflect.ValueOf( viewed_users[0] ) , &csv_headers , nil )
// func extract_fields( prefix string , v reflect.Value , fieldNames *[]string , fieldValues *[]string ) {
// 	t := v.Type()
// 	for i := 0; i < t.NumField(); i++ {
// 		field := t.Field(i)
// 		value := v.Field(i)
// 		fullName := fmt.Sprintf("%s%s", prefix, field.Name)
// 		if field.Type.Kind() == reflect.Struct {
// 			extract_fields(fullName+".", value, fieldNames, fieldValues)
// 		} else {
// 			if fieldNames != nil {
// 				*fieldNames = append(*fieldNames, fullName)
// 			}
// 			if fieldValues != nil {
// 				*fieldValues = append(*fieldValues, fmt.Sprint(value.Interface()))
// 			}
// 		}
// 	}
// }

func its( i int ) ( s string ) {
	s = strconv.Itoa( i )
	return
}

func GetReportMain( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	var csv_lines [][]string
	csv_headers := []string{
		"First Name" , "Middle Name" , "Last Name" ,
		"Email Address" , "Phone" ,
		"Primary Barcode" , "UUID" , "Spanish" ,
		"Street Number" , "Street Name" , "Address-2" , "City" , "State" , "ZipCode" ,
		"DOB - Month" , "DOB - Day" , "DOB - Year" ,
		"Family Size" , "Created Date" , "Created Time" , "Total Check-Ins" ,
		"Available - Tops" , "Available - Bottoms" , "Available - Dresses" , "Available - Shoes" , "Available - Seasonals" , "Available - Accessories" ,
		"Recieved - Tops" , "Recieved - Bottoms" , "Recieved - Dresses" , "Recieved - Shoes" , "Recieved - Seasonals" , "Recieved - Accessories" ,
	}
	csv_lines = append( csv_lines , csv_headers )

	// Extract Each User's Info into out custom csv structure
	// var viewed_users []user.User
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			is_spanish := strconv.FormatBool( viewed_user.Spanish )
			primary_barcode := ""
			if len( viewed_user.Barcodes ) > 0 { primary_barcode = viewed_user.Barcodes[ 0 ] }
			x_user_csv := []string{
				viewed_user.Identity.FirstName , viewed_user.Identity.MiddleName , viewed_user.Identity.LastName ,
				viewed_user.EmailAddress , viewed_user.PhoneNumber ,
				primary_barcode , viewed_user.UUID , is_spanish ,
				viewed_user.Identity.Address.StreetNumber , viewed_user.Identity.Address.StreetName , viewed_user.Identity.Address.AddressTwo , viewed_user.Identity.Address.City , viewed_user.Identity.Address.State , viewed_user.Identity.Address.ZipCode ,
				viewed_user.Identity.DateOfBirth.Month , its( viewed_user.Identity.DateOfBirth.Day ) , its( viewed_user.Identity.DateOfBirth.Year ) ,
				its( viewed_user.FamilySize ) , viewed_user.CreatedDate , viewed_user.CreatedTime , its( len( viewed_user.CheckIns ) ) ,
				its( viewed_user.Balance.General.Tops.Available ) , its( viewed_user.Balance.General.Bottoms.Available ) , its( viewed_user.Balance.General.Dresses.Available ) , its( viewed_user.Balance.Shoes.Available ) , its( viewed_user.Balance.Seasonals.Available ) , its( viewed_user.Balance.Accessories.Available ) ,
				its( viewed_user.Balance.General.Tops.Used ) , its( viewed_user.Balance.General.Bottoms.Used ) , its( viewed_user.Balance.General.Dresses.Used ) , its( viewed_user.Balance.Shoes.Used ) , its( viewed_user.Balance.Seasonals.Used ) , its( viewed_user.Balance.Accessories.Used ) ,
			}
			csv_lines = append( csv_lines , x_user_csv )
			return nil
		})
		return nil
	})

	// Build CSV File
	var csv_buffer bytes.Buffer
	writer := csv.NewWriter( &csv_buffer )
	for _ , row := range csv_lines {
		writer.Write( row )
	}
	writer.Flush()
	csv_bytes := csv_buffer.Bytes()
	context.Set( "Content-Type", "text/csv" )
	context.Set( "Content-Disposition", "attachment;filename=masters_closet_users.csv" )
	context.Set( "Content-Length", strconv.Itoa( len( csv_bytes ) ) )
	log.PrintlnConsole( "Downloaded Report-Main" , strconv.Itoa( len( csv_bytes ) ) )
	return context.Send( csv_bytes )

}
