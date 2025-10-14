package server

import (
	// "fmt"
	fiber "github.com/gofiber/fiber/v2"
	// types "github.com/0187773933/MastersCloset/v1/types"
	// bolt "github.com/boltdb/bolt"
	// logger "github.com/0187773933/MastersCloset/v1/logger"
)

var admin_ui_html_pages = map[ string ]string {
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

func ( s *Server ) RegisterAdminRoutes() {
	admin_route_group := s.FiberApp.Group( "/admin" )

	// HTML UI Pages
	admin_route_group.Get( "/login" , ServeLoginPage )

	for url , _ := range admin_ui_html_pages {
		// fmt.Println( "Registering" , url )
		admin_route_group.Get( url , s.ServeAuthenticatedPage )
	}

	s.FiberApp.Get( "/cdn/api.js" , func( context *fiber.Ctx ) ( error ) {
		if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }
		return context.SendFile( "./v1/server/cdn/api.js" )
	})

	admin_route_group.Get( "/logout" , s.AdminLogout )
	admin_route_group.Post( "/login" , s.AdminLogin )
	admin_route_group.Get( "/logs/get/log-file-names" , s.GetLogFileNames )
	admin_route_group.Get( "/logs/get/:file_name" , s.GetLogFile )

	admin_route_group.Get( "/checkins/get/:date" , s.GetCheckinsDate )
	admin_route_group.Get( "/checkins/delete/:uuid/:ulid" , s.DeleteCheckIn )
	admin_route_group.Get( "/checkins/get/:uuid/:ulid" , s.GetCheckIn )
	admin_route_group.Post( "/checkins/edit/:uuid/:ulid" , s.EditCheckIn )

	admin_route_group.Post( "/user/new" , s.HandleNewUserJoin )
	admin_route_group.Post( "/user/similar" , s.HandleUserSimilar ) // finds similar users reports
	admin_route_group.Get( "/user/similar/o/:uuid" , s.HandleUserSimilarObjects ) // finds similar user objects
	admin_route_group.Post( "/user/edit" , s.HandleUserEdit )
	admin_route_group.Get( "/user/delete/:uuid" , s.DeleteUser )
	// admin_route_group.Get( "/user/check/username" , CheckIfFirstNameLastNameAlreadyExists )

	admin_route_group.Post( "/user/checkin/:uuid" , s.UserCheckIn )
	// admin_route_group.Get( "/user/checkin/:uuid/:verb" , UserCheckIn )
	admin_route_group.Get( "/user/checkin/test/:uuid" , s.UserCheckInTest )

	admin_route_group.Get( "/user/get/all" , s.GetAllUsers )
	admin_route_group.Get( "/user/get/all/checkins" , s.GetAllCheckIns )
	admin_route_group.Get( "/user/get/all/emails" , s.GetAllEmails )
	admin_route_group.Get( "/user/get/all/phone-numbers" , s.GetAllPhoneNumbers )
	admin_route_group.Get( "/user/get/all/barcodes" , s.GetAllBarcodes )
	admin_route_group.Get( "/user/get/:uuid" , s.GetUser )
	admin_route_group.Get( "/user/get/barcode/:barcode" , s.GetUserViaBarcode )
	admin_route_group.Get( "/user/get/ulid/:ulid" , s.GetUserViaULID )
	// admin_route_group.Get( "/user/get/checkins/:date" , GetCheckinsDate )

	admin_route_group.Get( "/user/search/username/:username" , s.UserSearch )
	admin_route_group.Get( "/user/search/username/fuzzy/:username" , s.UserSearchFuzzy )
	admin_route_group.Get( "/print-test" , s.PrintTest )
	admin_route_group.Post( "/print" , s.Print )
	admin_route_group.Post( "/print2" , s.PrintTwo )

	admin_route_group.Get( "/user/reports/main" , s.GetReportMain )
	admin_route_group.Get( "/user/reports/mail-chimp" , s.GetReportMailChimp )

	admin_route_group.Post( "/user/sms" , s.SMSUser )
	admin_route_group.Post( "/user/sms/all" , s.SMSAllUsers )
	admin_route_group.Post( "/user/email/all" , s.EmailAllUsers )
	admin_route_group.Post( "/user/email" , s.EmailUser )

	admin_route_group.Post( "/transcribe/base-user-structure" , s.AudioToBaseUserStructure )

}