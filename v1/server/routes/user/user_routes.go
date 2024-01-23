package userroutes

import (
	"fmt"
	"time"
	// "math/rand"
	"strings"
	// json "encoding/json"
	// net_url "net/url"
	fiber "github.com/gofiber/fiber/v2"
	rate_limiter "github.com/gofiber/fiber/v2/middleware/limiter"
	types "github.com/0187773933/MastersCloset/v1/types"
	// bolt "github.com/0187773933/MastersCloset/v1/bolt"
	// utils "github.com/0187773933/MastersCloset/v1/utils"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	log "github.com/0187773933/MastersCloset/v1/log"
	// bleve "github.com/blevesearch/bleve/v2"
)

var GlobalConfig *types.ConfigFile

var public_limiter = rate_limiter.New(rate_limiter.Config{
	Max:        30, // set a different rate limit for this route
	Expiration: 1 * time.Second,
	// your remaining configurations...
	KeyGenerator: func(c *fiber.Ctx) string {
		return c.Get("x-forwarded-for")
	},
	LimitReached: func(c *fiber.Ctx) error {
		ip_address := c.IP()
		log_message := fmt.Sprintf( "%s === %s === %s === PUBLIC RATE LIMIT REACHED !!!" , ip_address , c.Method() , c.Path() );
		log.PrintlnConsole( log_message )
		c.Set( "Content-Type" , "text/html" )
		return c.SendString( "<html><h1>loading ...</h1><script>setTimeout(function(){ window.location.reload(1); }, 6);</script></html>" )
	} ,
})

var user_creation_limiter = rate_limiter.New(rate_limiter.Config{
	Max:        1, // set a different rate limit for this route
	Expiration: 60 * time.Minute,
	// your remaining configurations...
	KeyGenerator: func(c *fiber.Ctx) string {
		return c.Get("x-forwarded-for")
	},
	LimitReached: func(c *fiber.Ctx) error {
		ip_address := c.IP()
		log_message := fmt.Sprintf( "%s === %s === %s === PUBLIC USER CREATION RATE LIMIT REACHED !!!" , ip_address , c.Method() , c.Path() );
		log.PrintlnConsole( log_message )
		c.Set("Content-Type", "text/html")
		return c.SendString("<html><h1>loading ...</h1><script>setTimeout(function(){ window.location.reload(1); }, 6);</script></html>")
	},
})


func RegisterRoutes( fiber_app *fiber.App , config *types.ConfigFile ) {
	GlobalConfig = config

	fiber_app.Get( "/" , public_limiter , RenderHomePage )
	fiber_app.Get( "/logo.png" , public_limiter , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/logo.png" ) } )
	fiber_app.Get( "/cdn/utils.js" , public_limiter , func( context *fiber.Ctx ) ( error ) { context.Set( "Cache-Control" , "public, max-age=1" ); return context.SendFile( "./v1/server/cdn/utils.js" ) } )
	fiber_app.Get( "/cdn/ui.js" , public_limiter , func( context *fiber.Ctx ) ( error ) { context.Set( "Cache-Control" , "public, max-age=1" ); return context.SendFile( "./v1/server/cdn/ui.js" ) } )
	fiber_app.Get( "/cdn/ui.css" , public_limiter , func( context *fiber.Ctx ) ( error ) { context.Set( "Cache-Control" , "public, max-age=1" ); return context.SendFile( "./v1/server/cdn/ui.css" ) } )
	fiber_app.Get( "/cdn/verified.png" , public_limiter , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/verified.png" ) } )
	fiber_app.Get( "/favicon.ico" , public_limiter , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/favicon.ico" ) } )
	fiber_app.Get( "/cdn/sodium.js" , public_limiter , func( context *fiber.Ctx ) ( error ) { return context.SendFile( "./v1/server/cdn/sodium.js" ) } )
	fiber_app.Get( "/cdn/protobuf.min.js" , public_limiter , func( context *fiber.Ctx ) ( error ) { context.Set( "Cache-Control" , "public, max-age=1" ); return context.SendFile( "./v1/server/cdn/utils.js" ) } )

	fiber_app.Get( "/join" , public_limiter , RenderJoinPage )
	// fiber_app.Get( "/join/display" , public_limiter , CheckInDisplay )
	// fiber_app.Post( "/user/new" , user_creation_limiter , HandleNewUserJoin )
	fiber_app.Get( "/checkin" , public_limiter , CheckIn )

	user_route_group := fiber_app.Group( "/user" )
	user_route_group.Get( "/login/fresh/:uuid" , public_limiter , LoginFresh )
	// user_route_group.Get( "/login/success/:uuid" , LoginSuccess )
	// user_route_group.Get( "/checkin/display/:uuid" , public_limiter , CheckInDisplay )
	user_route_group.Get( "/checkin" , public_limiter , CheckIn )
	user_route_group.Get( "/checkin/silent/:uuid" , public_limiter , CheckInSilentTest )

	fiber_app.Get( "/*" , public_limiter , func( context *fiber.Ctx ) ( error ) { return context.Redirect( "/" ) } )
}

func check_if_user_cookie_exists( context *fiber.Ctx ) ( result bool ) {
	result = false
	user_cookie := context.Cookies( "the-masters-closet-user" )
	if user_cookie == "" { return }
	user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )
	// result = uuid.IsUUID( user_cookie_value )
	if user_cookie_value != "" { result = true }
	return
}

func check_if_user_ulid_cookie_exists( context *fiber.Ctx ) ( result bool ) {
	fmt.Println( "check_if_user_ulid_cookie_exists()" )
	result = false
    cookieHeader := string(context.Request().Header.Peek("Cookie"))
    cookies := strings.Split(cookieHeader, "; ")
    fmt.Println( cookies )
	user_cookie := context.Cookies( "the-masters-closet-user-ulid=" )
	fmt.Println( user_cookie )
	if user_cookie == "" { return }
	fmt.Println( user_cookie )
	user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )
	fmt.Println( user_cookie_value )
	if user_cookie_value != "" { result = true }
	return
}

func check_if_admin_cookie_exists( context *fiber.Ctx ) ( result bool ) {
	result = false
	admin_cookie := context.Cookies( "the-masters-closet-admin" )
	if admin_cookie == "" { return }
	admin_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , admin_cookie )
	if admin_cookie_value != GlobalConfig.ServerCookieAdminSecretMessage { fmt.Println( "admin cookie secret message was not equal" ); return }
	result = true
	return
}

func serve_failed_check_in_attempt( context *fiber.Ctx ) ( error ) {
	// return context.Redirect( "/join" )
	context.Set( "Content-Type" , "text/html" )
	// return context.SendString( "<h1>check-in failed</h1>" )
	return context.SendFile( "./v1/server/html/user_check_in_failed.html" )
}

func RenderHomePage( context *fiber.Ctx ) ( error ) {
	fmt.Println( "RenderHomePage()" )
	context.Set( "Content-Type" , "text/html" )

	admin_logged_in := check_if_admin_cookie_exists( context )
	if admin_logged_in == true {
		fmt.Println( "RenderHomePage() --> Admin" )
		return context.SendFile( "./v1/server/html/admin.html" )
	}

	user_logged_in := check_if_user_cookie_exists( context )
	if user_logged_in == true {
		fmt.Println( "RenderHomePage() --> User UUID" )
		user_cookie := context.Cookies( "the-masters-closet-user" )
		user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )
		fmt.Println( "user logged in" , user_cookie_value )
		return context.SendFile( "./v1/server/html/user_home.html" )
	}

	user_ulid_cookie_exists := check_if_user_ulid_cookie_exists( context )
	if user_ulid_cookie_exists == true {
		fmt.Println( "RenderHomePage() --> User ULID" )
		user_ulid_cookie := context.Cookies( "the-masters-closet-user-ulid" )
		fmt.Println( user_ulid_cookie )
		user_ulid_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_ulid_cookie )
		fmt.Println( user_ulid_cookie_value )
		return context.SendFile( "./v1/server/html/user_home.html" )
	}

	fmt.Println( "RenderHomePage() --> Default" )
	return context.SendFile( "./v1/server/html/home.html" )
}

func RenderJoinPage( context *fiber.Ctx ) ( error ) {
	user_logged_in := check_if_user_cookie_exists( context )
	if user_logged_in == true {
		return context.SendFile( "./v1/server/html/user_home.html" )
	}
	return context.SendFile( "./v1/server/html/user_new.html" )
}


func LoginFresh( context *fiber.Ctx ) ( error ) {
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	x_user_uuid := context.Params( "uuid" )
	x_user := user.GetByUUID( x_user_uuid , db , GlobalConfig.BoltDBEncryptionKey )
	if x_user.UUID == "" {
		context.Set( "Content-Type" , "text/html" )
		return context.SendString( "<h1>Login Failed</h1>" )
	}

	// Manual Check In For First Time Login
	user.CheckInUser( x_user.UUID , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )

	context.Cookie(
		&fiber.Cookie{
			Name: "the-masters-closet-user" ,
			Value: encryption.SecretBoxEncrypt( GlobalConfig.BoltDBEncryptionKey , x_user.UUID ) ,
			Secure: true , // dev
			Path: "/" , // fucking webkit
			// Domain: "blah.ngrok.io" , // probably should set this for webkit
			HTTPOnly: true ,
			SameSite: "Lax" ,
			Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
		} ,
	)

	return context.SendFile( "./v1/server/html/user_login_success.html" )
}

func CheckInSilentTest( context *fiber.Ctx ) ( error ) {
	x_user_uuid := context.Params( "uuid" )
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	check_in_result , milliseconds_remaining , balance , name_string , family_size := user.CheckInTest( x_user_uuid , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )
	return context.JSON( fiber.Map{
		"route": "/user/checkin/silent/:uuid" ,
		"result": fiber.Map{
			"check_in_possible": check_in_result ,
			"milliseconds_remaining": milliseconds_remaining ,
			"balance": balance ,
			"name_string": name_string ,
			"family_size": family_size ,
		} ,
	})
}

// https://docs.gofiber.io/api/ctx#cookie
// http://localhost:5950/user/new/:username
func CheckIn( context *fiber.Ctx ) ( error ) {

	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	// validate they have a stored user cookie
	user_cookie := context.Cookies( "the-masters-closet-user" )
	// if user_cookie == "" { fmt.Println( "user cookie was blank" ); return serve_failed_check_in_attempt( context ) }
	if user_cookie == "" { fmt.Println( "user cookie was blank" ); return context.SendFile( "./v1/server/html/user_new.html" ) }
	user_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , user_cookie )
	x_user := user.GetByUUID( user_cookie_value , db , GlobalConfig.BoltDBEncryptionKey )
	// if x_user.UUID == "" { fmt.Println( "UUID stored in user cookie was blank" ); return serve_failed_check_in_attempt( context ) }
	if x_user.UUID == "" { fmt.Println( "UUID stored in user cookie was blank" ); return context.SendFile( "./v1/server/html/user_new.html" ) }
	// TODO : render a different page if the check-in would fail ?
	// check_in_test , time_remaining := user.CheckInUser( x_user.UUID , db , GlobalConfig.BoltDBEncryptionKey , GlobalConfig.CheckInCoolOffDays )
	// fmt.Println( "Pre-Check-In Test Result ===" , check_in_test , "Time Remaining ===" , time_remaining )
	return context.Redirect( fmt.Sprintf( "/user/checkin/display/%s" , x_user.UUID ) )
}

func CheckInDisplay( context *fiber.Ctx ) ( error ) {
	return context.SendFile( "./v1/server/html/user_check_in.html" )
}

// func HandleNewUserJoin( context *fiber.Ctx ) ( error ) {

// 	var viewed_user user.User
// 	json.Unmarshal( context.Body() , &viewed_user )

// 	viewed_user.Config = GlobalConfig

// 	// Treat this as a Temp User
// 	if viewed_user.Identity.FirstName == "" && viewed_user.Identity.MiddleName == "" && viewed_user.Identity.LastName == "" {
// 		viewed_user.Identity.FirstName = "Temp"
// 		rand.Seed( time.Now().UnixNano() )
// 		viewed_user.Identity.MiddleName = fmt.Sprintf( "%v%v%v%v%v%v" , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) )
// 		viewed_user.Identity.LastName = fmt.Sprintf( "%v%v%v%v%v%v" , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) )
// 	}

// 	viewed_user.FormatUsername()

// 	new_user := user.New( viewed_user.Username , GlobalConfig )
// 	log.Println( new_user )

// 	viewed_user.UUID = new_user.UUID
// 	viewed_user.CreatedDate = new_user.CreatedDate
// 	viewed_user.CreatedTime = new_user.CreatedTime
// 	viewed_user.Save()

// 	// add to search index
// 	search_index , _ := bleve.Open( GlobalConfig.BleveSearchPath )
// 	defer search_index.Close()
// 	search_item := types.SearchItem{
// 		UUID: new_user.UUID ,
// 		Name: viewed_user.NameString ,
// 	}
// 	log.Printf( "Updating Search Index with : %s\n" , viewed_user.NameString )
// 	search_index.Index( new_user.UUID , search_item )

// 	log.PrintlnConsole( "Created New User :" , viewed_user.NameString )

// 	context.Cookie(
// 		&fiber.Cookie{
// 			Name: "the-masters-closet-user" ,
// 			Value: encryption.SecretBoxEncrypt( GlobalConfig.BoltDBEncryptionKey , viewed_user.UUID ) ,
// 			Secure: true ,
// 			Path: "/" ,
// 			// Domain: "blah.ngrok.io" , // probably should set this for webkit
// 			HTTPOnly: true ,
// 			SameSite: "Lax" ,
// 			Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
// 		} ,
// 	)

// 	// viewed_user.Save();
// 	return context.JSON( fiber.Map{
// 		"route": "/user/new" ,
// 		"result": viewed_user.UUID ,
// 	})
// }