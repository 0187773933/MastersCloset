package adminroutes

import (
	// "fmt"
	fiber "github.com/gofiber/fiber/v2"
	utils "github.com/0187773933/MastersCloset/v1/utils"
)

// TODO = fix , because we finally fixed logs

func GetLogFileNames( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	return context.JSON( fiber.Map{
		"route": "/admin/logs/get-log-file-names" ,
		"result": utils.GetLogFileNames() ,
	})
}

func GetLogFile( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	file_path := context.Params( "file_name" )
	return context.JSON( fiber.Map{
		"route": "/admin/logs/:file_name" ,
		"file_path": file_path ,
		"result": utils.GetLogFile( file_path ) ,
	})
}