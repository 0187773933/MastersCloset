package adminroutes

import (
	// "fmt"
	fiber "github.com/gofiber/fiber/v2"
	log "github.com/0187773933/MastersCloset/v1/log"
)

func GetLogFileNames( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	return context.JSON( fiber.Map{
		"route": "/admin/logs/get-log-file-names" ,
		"result": log.GetLogFileNames() ,
	})
}

func GetLogFile( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	file_path := context.Params( "file_name" )
	return context.JSON( fiber.Map{
		"route": "/admin/logs/:file_name" ,
		"file_path": file_path ,
		"result": log.GetLogFile( file_path ) ,
	})
}