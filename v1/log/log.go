package log

import (
	// "io"
	"os"
	"bufio"
	"time"
	"fmt"
	"log"
	"io/ioutil"
	"strings"
	"sort"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	// logrus "github.com/sirupsen/logrus"
	types "github.com/0187773933/MastersCloset/v1/types"
	utils "github.com/0187773933/MastersCloset/v1/utils"
)
// var log Logger
// type Logger struct {
// 	Logrus *logrus.Logger
// }


func GetLogFileNames() ( result []string ) {
	files , _ := ioutil.ReadDir( "./logs" )
	for _ , file := range files {
		if file.IsDir() { continue; }
		file_name := file.Name()
		if strings.HasSuffix( file_name , ".log" ) {
			// fmt.Println( file_name )
			// result = append( result , strings.Split( file_name , "-" )[ 0 ] )
			result = append( result , file_name )
		}
	}
	sort.Strings( result )
	return
}

func GetLogFile( file_path string ) ( result []string ) {
	file , _ := os.Open( fmt.Sprintf( "./logs/%s" , file_path ) )
	defer file.Close()
	scanner := bufio.NewScanner( file )
	for scanner.Scan() {
		result = append( result , scanner.Text() )
	}
	return result
}

var ROTATOR *lumberjack.Logger
var finger_print string

func Init( config types.ConfigFile ) {
	finger_print = config.FingerPrint
	prepended_timestamp := time.Now().Format( "20060102" )
	// log.SetFlags( 0 )
	ROTATOR = &lumberjack.Logger{
		Filename: fmt.Sprintf( "./logs/%s-%s.log" , prepended_timestamp , config.FingerPrint ) ,
		MaxSize: 100 , // megabytes
		// MaxBackups: 3 ,   // number of backups
		MaxAge: 1 , // days
		Compress: true , // compress the rotated log files
	}
	log.SetFlags( 0 )
	log.SetOutput( ROTATOR )
}

func Close() {
	ROTATOR.Close()
}

func Println( args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	args = append( []interface{}{ time_string , "===" , finger_print , "===" } , args... )
	// fields := logrus.Fields{ "time": time_string , }
	// log.Logrus.WithFields( fields ).Println( args... )
	// log.Logrus.Println( args... )
	log.Println( args... )
}

func PrintlnConsole(args ...interface{}) {
	if len(args) < 1 { return }
	time_string := utils.GetFormattedTimeString()
	final_msg := make( []interface{} , 0 )
	final_msg = append( final_msg , time_string , "===" , finger_print , "===" )
	final_msg = append( final_msg , args... )
	log.Println( final_msg... )
	fmt.Println( final_msg... )
}

func Printf( format_string string , args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	sent_format := fmt.Sprintf( format_string , args... )
	log.Printf( "%s === %s === %s" , time_string , finger_print , sent_format )
}

func PrintfConsole( format_string string , args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	sent_format := fmt.Sprintf( format_string , args... )
	log.Printf( "%s === %s === %s" , time_string , finger_print , sent_format )
	fmt.Printf( "%s === %s === %s" , time_string , finger_print , sent_format )
}

func Debug( args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	args = append( []interface{}{ time_string , "===" , finger_print , "===" , "DEBUG" , "===" } , args... )
	// fields := logrus.Fields{ "time": time_string , }
	// log.Logrus.WithFields( fields ).Debug( args... )
	// log.Logrus.Debug( args... )
	log.Println( args... )
}

func Info( args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	args = append( []interface{}{ time_string , "===" , finger_print , "===" , "DEBUG" , "===" } , args... )
	// fields := logrus.Fields{ "time": time_string , }
	// log.Logrus.WithFields( fields ).Debug( args... )
	// log.Logrus.Debug( args... )
	log.Println( args... )
}

func Error( args ...interface{} ) {
	time_string := utils.GetFormattedTimeString()
	args = append( []interface{}{ "ERROR !!!" , time_string , "===" , finger_print , "===" } , args... )
	// fields := logrus.Fields{ "time": time_string , }
	// log.Logrus.WithFields( fields ).Debug( args... )
	// log.Logrus.Debug( args... )
	log.Println( args... )
	fmt.Println( args... )
}
