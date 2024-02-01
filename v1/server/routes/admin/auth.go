package adminroutes

import (
	// "fmt"
	"time"
	fiber "github.com/gofiber/fiber/v2"
	bcrypt "golang.org/x/crypto/bcrypt"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	// log "github.com/0187773933/MastersCloset/v1/log"
)

func validate_login_credentials( context *fiber.Ctx ) ( result bool ) {
	result = false
	uploaded_username := context.FormValue( "username" )
	if uploaded_username == "" { log.Println( "username empty" ); return }
	if uploaded_username != GlobalConfig.AdminUsername { log.Println( "username not correct" ); return }
	uploaded_password := context.FormValue( "password" )
	if uploaded_password == "" { log.Println( "password empty" ); return }
	password_matches := bcrypt.CompareHashAndPassword( []byte( uploaded_password ) , []byte( GlobalConfig.AdminPassword ) )
	if password_matches != nil { log.Println( "bcrypted password doesn't match" ); return }
	result = true
	return
}

func Logout( context *fiber.Ctx ) ( error ) {
	context.Cookie( &fiber.Cookie{
		Name: "the-masters-closet-admin" ,
		Value: "" ,
		Expires: time.Now().Add( -time.Hour ) , // set the expiration to the past
		HTTPOnly: true ,
		Secure: true ,
	})
	context.Set( "Content-Type" , "text/html" )
	return context.SendString( "<h1>Logged Out</h1>" )
}

// POST http://localhost:5950/admin/login
func HandleLogin( context *fiber.Ctx ) ( error ) {
	valid_login := validate_login_credentials( context )
	if valid_login == false { return serve_failed_attempt( context ) }
	context.Cookie(
		&fiber.Cookie{
			Name: "the-masters-closet-admin" ,
			Value: encryption.SecretBoxEncrypt( GlobalConfig.BoltDBEncryptionKey , GlobalConfig.ServerCookieAdminSecretMessage ) ,
			Secure: true ,
			Path: "/" ,
			// Domain: "blah.ngrok.io" , // probably should set this for webkit
			HTTPOnly: true ,
			SameSite: "Lax" ,
			Expires: time.Now().AddDate( 10 , 0 , 0 ) , // aka 10 years from now
		} ,
	)
	return context.Redirect( "/" )
}

func validate_admin_cookie( context *fiber.Ctx ) ( result bool ) {
	result = false
	admin_cookie := context.Cookies( "the-masters-closet-admin" )
	if admin_cookie == "" { log.Debug( "admin cookie was blank" ); return }
	admin_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , admin_cookie )
	if admin_cookie_value != GlobalConfig.ServerCookieAdminSecretMessage { log.Debug( "admin cookie secret message was not equal" ); return }
	result = true
	return
}

func validate_admin_session( context *fiber.Ctx ) ( result bool ) {
	result = false
	admin_cookie := context.Cookies( "the-masters-closet-admin" )
	if admin_cookie != "" {
		admin_cookie_value := encryption.SecretBoxDecrypt( GlobalConfig.BoltDBEncryptionKey , admin_cookie )
		if admin_cookie_value == GlobalConfig.ServerCookieAdminSecretMessage {
			result = true
			return
		}
	}
	admin_api_key_header := context.Get( "key" )
	if admin_api_key_header != "" {
		if admin_api_key_header == GlobalConfig.ServerAPIKey {
			result = true
			return
		}
	}
	admin_api_key_query := context.Query( "k" )
	if admin_api_key_query != "" {
		if admin_api_key_query == GlobalConfig.ServerAPIKey {
			result = true
			return
		}
	}
	return
}
