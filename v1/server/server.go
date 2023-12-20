package server

import (
	"fmt"
	fiber "github.com/gofiber/fiber/v2"
	fiber_cookie "github.com/gofiber/fiber/v2/middleware/encryptcookie"
	fiber_cors "github.com/gofiber/fiber/v2/middleware/cors"
	// favicon "github.com/gofiber/fiber/v2/middleware/favicon"
	// try "github.com/manucorporat/try"
	types "github.com/0187773933/MastersCloset/v1/types"
	user_routes "github.com/0187773933/MastersCloset/v1/server/routes/user"
	admin_routes "github.com/0187773933/MastersCloset/v1/server/routes/admin"
	// "os"
	log "github.com/0187773933/MastersCloset/v1/log"
)

var GlobalConfig *types.ConfigFile

type Server struct {
	FiberApp *fiber.App `json:"fiber_app"`
	Config types.ConfigFile `json:"config"`
}

func request_logging_middleware( context *fiber.Ctx ) ( error ) {
	ip_address := context.Get( "x-forwarded-for" )
	if ip_address == "" { ip_address = context.IP() }
	// log_message := fmt.Sprintf( "%s === %s === %s === %s === %s" , time_string , GlobalConfig.FingerPrint , ip_address , context.Method() , context.Path() )
	// log_message := fmt.Sprintf( "%s === %s === %s === %s" , time_string , GlobalConfig.FingerPrint , context.Method() , context.Path() )
	// log_message := fmt.Sprintf( "%s === %s" , context.Method() , context.Path() )
	log_message := fmt.Sprintf( "%s === %s === %s" , ip_address , context.Method() , context.Path() );
	log.Println( log_message )
	return context.Next()
}

func New( config types.ConfigFile ) ( server Server ) {

	server.FiberApp = fiber.New()
	server.Config = config
	GlobalConfig = &config

	server.FiberApp.Use( request_logging_middleware )
	// server.FiberApp.Use( favicon.New( favicon.Config{
	// 	File: "./v1/server/cdn/favicon.ico" ,
	// }))

	// temp_key := fiber_cookie.GenerateKey()
	// fmt.Println( temp_key )
	server.FiberApp.Use( fiber_cookie.New( fiber_cookie.Config{
		Key: server.Config.ServerCookieSecret ,
	}))

	server.FiberApp.Use( fiber_cors.New( fiber_cors.Config{
		AllowOrigins: fmt.Sprintf( "%s, %s" , server.Config.ServerBaseUrl , server.Config.ServerLiveUrl ) ,
		AllowHeaders:  "Origin, Content-Type, Accept, key",
	}))

	server.SetupRoutes()
	// server.FiberApp.Get( "/*" , func( context *fiber.Ctx ) ( error ) { return context.Redirect( "/" ) } )
	return
}

func ( s *Server ) SetupRoutes() {
	admin_routes.RegisterRoutes( s.FiberApp , &s.Config )
	user_routes.RegisterRoutes( s.FiberApp , &s.Config )
}

func ( s *Server ) Start() {
	fmt.Println( "\n" )
	log.PrintfConsole( "Listening on http://localhost:%s\n" , s.Config.ServerPort )
	fmt.Printf( "Admin Login @ http://localhost:%s/admin/login\n" , s.Config.ServerPort )
	fmt.Printf( "Admin Login @ %s/admin/login\n" , s.Config.ServerLiveUrl )
	fmt.Printf( "Admin Username === %s\n" , s.Config.AdminUsername )
	fmt.Printf( "Admin Password === %s\n" , s.Config.AdminPassword )
	s.FiberApp.Listen( fmt.Sprintf( ":%s" , s.Config.ServerPort ) )
}

