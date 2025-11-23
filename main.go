package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"strings"
	// "time"
	"path/filepath"
	"context"
	bolt "github.com/boltdb/bolt"
	server "github.com/0187773933/MastersCloset/v1/server"
	utils "github.com/0187773933/MastersCloset/v1/utils"
	remotesync "github.com/0187773933/MastersCloset/v1/remotesync"
	// log "github.com/0187773933/MastersCloset/v1/log"
	logger "github.com/0187773933/MastersCloset/v1/logger"
	types "github.com/0187773933/MastersCloset/v1/types"
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

func FixDBAndSearchIndex( config *types.ConfigFile ) {
	home_dir , _ := os.UserHomeDir()
	dest_dir := filepath.Join( home_dir , ".config" , "mct" )
	os.MkdirAll( dest_dir , 0755 )

	new_bolt_db_name := strings.TrimSuffix( config.BoltDBPath , ".db" ) + "_" + config.FingerPrint + ".db"
	new_bolt_db_file_path := filepath.Join( dest_dir , new_bolt_db_name )
	// new_bolt_db_file_path := new_bolt_db_name
	_ , new_bolt_err := os.Stat( new_bolt_db_file_path )
	if os.IsNotExist( new_bolt_err ) {
		fmt.Println( "Local Bolt DB Didn't Exist ! Making Copy from Dropbox" )
		utils.CopyFile( config.BoltDBPath , new_bolt_db_file_path )
	}
	config.BoltDBPath = new_bolt_db_file_path

	new_bleve_name := strings.TrimSuffix( config.BleveSearchPath , ".bleve" ) + "_" + config.FingerPrint + ".bleve"
	new_bleve_path := filepath.Join( dest_dir , new_bleve_name )
	// new_bleve_path := new_bleve_name
	_ , new_bleve_err := os.Stat( new_bleve_path );
	if os.IsNotExist( new_bleve_err ) {
		fmt.Println( "Local Bleve Search Index Didn't Exist ! Making Copy from Dropbox" )
		utils.CopyDir( config.BleveSearchPath , new_bleve_path )
	}
	config.BleveSearchPath = new_bleve_path

}

func main() {
	// utils.GenerateNewKeys()

	config_file_path , _ := filepath.Abs( os.Args[ 1 ] )
	config := utils.ParseConfig( config_file_path )
	config.FingerPrint = utils.FingerPrintPassive()

	logger.Init()
	logger.Log.Info( "Hola , Christ Lives" )
	logger.Log.Debug( fmt.Sprintf( "Loaded Config File From : %s\n" , config_file_path ) )
	SetupCloseHandler()

	// patch/fix local db and bleve index
	FixDBAndSearchIndex( &config )
	fmt.Println( config.BoltDBPath )
	fmt.Println( config.BleveSearchPath )

	// log.Init( config )
	utils.WriteJS_API( config.ServerLiveUrl , config.ServerAPIKey , config.LocalHostUrl )
	db , db_err := bolt.Open( config.BoltDBPath , 0600 , &bolt.Options{} )
	// db , db_err := bolt.Open( config.BoltDBPath , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
	if db_err != nil { logger.Log.Fatal( db_err.Error() ) }

	s = server.New( config , db )
	rsctx , rsctxcancel = context.WithCancel( context.Background() )
	rs = remotesync.New( db , rsctx , &config )
	rs.Start()
	s.Start()
}