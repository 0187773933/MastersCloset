package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"path/filepath"
	server "github.com/0187773933/MastersCloset/v1/server"
	utils "github.com/0187773933/MastersCloset/v1/utils"
	log "github.com/0187773933/MastersCloset/v1/log"
	logger "github.com/0187773933/MastersCloset/v1/logger"
)

var s server.Server

func SetupCloseHandler() {
	c := make( chan os.Signal )
	signal.Notify( c , os.Interrupt , syscall.SIGTERM , syscall.SIGINT )
	go func() {
		<-c
		fmt.Println( "\r- Ctrl+C pressed in Terminal" )
		log.PrintlnConsole( "Shutting Down Master's Closet Tracking Server" )
		utils.WriteJS_API( "" , "" , "" )
		s.FiberApp.Shutdown()
		log.Close()
		utils.WriteJS_API( "" , "" , "" )
		os.Exit( 0 )
	}()
}

func main() {
	// utils.GenerateNewKeys()
	logger.Init()
	logger.Log.Printf( "Hola" )
	SetupCloseHandler()
	config_file_path , _ := filepath.Abs( os.Args[ 1 ] )
	fmt.Printf( "Loaded Config File From : %s\n" , config_file_path )
	config := utils.ParseConfig( config_file_path )
	config.FingerPrint = utils.FingerPrint( &config )
	fmt.Println( config )
	log.Init( config )
	utils.WriteJS_API( config.ServerLiveUrl , config.ServerAPIKey , config.LocalHostUrl )
	s = server.New( config )
	s.Start()
}