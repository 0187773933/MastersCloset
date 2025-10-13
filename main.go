package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"time"
	"path/filepath"
	"context"
	bolt "github.com/boltdb/bolt"
	server "github.com/0187773933/MastersCloset/v1/server"
	utils "github.com/0187773933/MastersCloset/v1/utils"
	remotesync "github.com/0187773933/MastersCloset/v1/remotesync"
	// log "github.com/0187773933/MastersCloset/v1/log"
	logger "github.com/0187773933/MastersCloset/v1/logger"
)

var s server.Server
var db *bolt.DB
var rs remotesync.RemoteSync
var rsctx context.Context
var rsctxcancel context.CancelFunc

func SetupCloseHandler() {
	c := make( chan os.Signal )
	signal.Notify( c , os.Interrupt , syscall.SIGTERM , syscall.SIGINT )
	go func() {
		<-c
		fmt.Println( "\r- Ctrl+C pressed in Terminal" )
		logger.Log.Info( "Shutting Down Master's Closet Tracking Server" )
		utils.WriteJS_API( "" , "" , "" )
		s.FiberApp.Shutdown()
		// log.Close()
		db.Close()
		rsctxcancel()
		utils.WriteJS_API( "" , "" , "" )
		os.Exit( 0 )
	}()
}

func main() {
	// utils.GenerateNewKeys()
	logger.Init()
	logger.Log.Info( "Hola" )
	SetupCloseHandler()
	config_file_path , _ := filepath.Abs( os.Args[ 1 ] )
	logger.Log.Debug( fmt.Sprintf( "Loaded Config File From : %s\n" , config_file_path ) )
	config := utils.ParseConfig( config_file_path )
	config.FingerPrint = utils.FingerPrint( &config )
	fmt.Println( config )
	// log.Init( config )
	utils.WriteJS_API( config.ServerLiveUrl , config.ServerAPIKey , config.LocalHostUrl )
	s = server.New( config )
	db , _ := bolt.Open( config.BoltDBPath , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
	s.DB = db
	rsctx , rsctxcancel = context.WithCancel( context.Background() )
	rs = remotesync.New( db , rsctx , &config )
	rs.Start()
	s.Start()
}