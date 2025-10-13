package adminroutes

import (
	// "fmt"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersCloset/v1/types"
	bolt "github.com/boltdb/bolt"
	logger "github.com/0187773933/MastersCloset/v1/logger"
)

var GlobalConfig *types.ConfigFile
var log = logger.GetLogger()

var ui_html_pages = map[ string ]string {
	"/": "./v1/server/html/admin.html" ,
	"/users": "./v1/server/html/admin_view_users.html" ,
	"/user/new": "./v1/server/html/admin_user_new.html" ,
	"/user/new/:uuid": "./v1/server/html/admin_user_new.html" ,
	"/user/new/:uuid/edit": "./v1/server/html/admin_user_new.html" ,
	"/user/new/handoff/:uuid": "./v1/server/html/admin_user_new_handoff.html" ,
	"/user/checkin": "./v1/server/html/admin_user_checkin.html" ,
	"/user/checkin/:uuid": "./v1/server/html/admin_user_checkin.html" ,
	"/user/checkin/:uuid/edit": "./v1/server/html/admin_user_checkin.html" ,
	"/user/checkin/:uuid/edit/:ulid": "./v1/server/html/admin_edit_checkin.html" ,
	"/user/checkin/new": "./v1/server/html/admin_user_checkin.html" ,
	"/user/edit/:uuid": "./v1/server/html/admin_user_edit.html" ,
	"/user/sms/:uuid": "./v1/server/html/admin_sms_user.html" ,
	"/user/email/:uuid": "./v1/server/html/admin_email_user.html" ,
	"/checkins": "./v1/server/html/admin_view_total_checkins.html" ,
	"/checkins/:date": "./v1/server/html/admin_view_checkins.html" ,
	"/emails": "./v1/server/html/admin_view_all_emails.html" ,
	"/phone-numbers": "./v1/server/html/admin_view_all_phone_numbers.html" ,
	"/barcodes": "./v1/server/html/admin_view_all_barcodes.html" ,
	"/sms": "./v1/server/html/admin_sms_all_users.html" ,
	"/email": "./v1/server/html/admin_email_all_users.html" ,
	"/logs": "./v1/server/html/admin_view_all_log_files.html" ,
	"/logs/:file_name": "./v1/server/html/admin_view_log_file.html" ,
	"/at": "./v1/server/html/audio_test.html" ,
}

func _get_db( context *fiber.Ctx ) ( *bolt.DB ) {
	db := context.Locals( "db" )
	if db == nil {
		log.Error( "Database not found in context locals" )
		return nil
	}
	db_cast , ok := db.( *bolt.DB )
	if !ok {
		log.Error( "Database in context locals is not of type *bolt.DB" )
		return nil
	}
	return db_cast
}

func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile , db *bolt.DB ) {
	GlobalConfig = config
	admin_route_group := fiber_app.Group( "/admin" )

	// HTML UI Pages
	admin_route_group.Get( "/login" , ServeLoginPage )

	for url , _ := range ui_html_pages {
		// fmt.Println( "Registering" , url )
		admin_route_group.Get( url , ServeAuthenticatedPage )
	}

	fiber_app.Get( "/cdn/api.js" , func( context *fiber.Ctx ) ( error ) {
		if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
		return context.SendFile( "./v1/server/cdn/api.js" )
	})

	admin_route_group.Use( func( c *fiber.Ctx ) error { c.Locals( "db" , db ); return c.Next() } )

	admin_route_group.Get( "/logout" , Logout )
	admin_route_group.Post( "/login" , HandleLogin )
	admin_route_group.Get( "/logs/get/log-file-names" , GetLogFileNames )
	admin_route_group.Get( "/logs/get/:file_name" , GetLogFile )

	admin_route_group.Get( "/checkins/get/:date" , GetCheckinsDate )
	admin_route_group.Get( "/checkins/delete/:uuid/:ulid" , DeleteCheckIn )
	admin_route_group.Get( "/checkins/get/:uuid/:ulid" , GetCheckIn )
	admin_route_group.Post( "/checkins/edit/:uuid/:ulid" , EditCheckIn )

	admin_route_group.Post( "/user/new" , HandleNewUserJoin )
	admin_route_group.Post( "/user/similar" , HandleUserSimilar ) // finds similar users reports
	admin_route_group.Get( "/user/similar/o/:uuid" , HandleUserSimilarObjects ) // finds similar user objects
	admin_route_group.Post( "/user/edit" , HandleUserEdit )
	admin_route_group.Get( "/user/delete/:uuid" , DeleteUser )
	// admin_route_group.Get( "/user/check/username" , CheckIfFirstNameLastNameAlreadyExists )

	admin_route_group.Post( "/user/checkin/:uuid" , UserCheckIn )
	// admin_route_group.Get( "/user/checkin/:uuid/:verb" , UserCheckIn )
	admin_route_group.Get( "/user/checkin/test/:uuid" , UserCheckInTest )

	admin_route_group.Get( "/user/get/all" , GetAllUsers )
	admin_route_group.Get( "/user/get/all/checkins" , GetAllCheckIns )
	admin_route_group.Get( "/user/get/all/emails" , GetAllEmails )
	admin_route_group.Get( "/user/get/all/phone-numbers" , GetAllPhoneNumbers )
	admin_route_group.Get( "/user/get/all/barcodes" , GetAllBarcodes )
	admin_route_group.Get( "/user/get/:uuid" , GetUser )
	admin_route_group.Get( "/user/get/barcode/:barcode" , GetUserViaBarcode )
	admin_route_group.Get( "/user/get/ulid/:ulid" , GetUserViaULID )
	// admin_route_group.Get( "/user/get/checkins/:date" , GetCheckinsDate )

	admin_route_group.Get( "/user/search/username/:username" , UserSearch )
	admin_route_group.Get( "/user/search/username/fuzzy/:username" , UserSearchFuzzy )
	admin_route_group.Get( "/print-test" , PrintTest )
	admin_route_group.Post( "/print" , Print )
	admin_route_group.Post( "/print2" , PrintTwo )

	admin_route_group.Get( "/user/reports/main" , GetReportMain )
	admin_route_group.Get( "/user/reports/mail-chimp" , GetReportMailChimp )

	admin_route_group.Post( "/user/sms" , SMSUser )
	admin_route_group.Post( "/user/sms/all" , SMSAllUsers )
	admin_route_group.Post( "/user/email/all" , EmailAllUsers )
	admin_route_group.Post( "/user/email" , EmailUser )

	admin_route_group.Post( "/transcribe/base-user-structure" , AudioToBaseUserStructure )

}