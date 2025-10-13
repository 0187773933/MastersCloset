package utils

import (
	"os"
	"os/user"
	"embed"
	"runtime"
	"bufio"
	"time"
	"net"
	"fmt"
	// "io"
	sha256 "crypto/sha256"
	hex "encoding/hex"
	// index_sort "github.com/mkmik/argsort"
	"sort"
	"strconv"
	"strings"
	json "encoding/json"
	"unicode"
	"io/ioutil"
	types "github.com/0187773933/MastersCloset/v1/types"
	fiber "github.com/gofiber/fiber/v2"
	fiber_cookie "github.com/gofiber/fiber/v2/middleware/encryptcookie"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	cpu "github.com/shirou/gopsutil/cpu"
	bolt_api "github.com/boltdb/bolt"
	// uuid "github.com/satori/go.uuid"
)

//go:embed zoneinfo
var ZoneInfoFS embed.FS
var location = get_location( "America/New_York" )

func ParseConfig( file_path string ) ( result types.ConfigFile ) {
	file_data , _ := ioutil.ReadFile( file_path )
	err := json.Unmarshal( file_data , &result )
	if err != nil { fmt.Println( err ) }
	return
}

// https://stackoverflow.com/a/28862477
func GetLocalIPAddresses() ( ip_addresses []string ) {
	host , _ := os.Hostname()
	addrs , _ := net.LookupIP( host )
	encountered := make( map[ string ]bool )
	for _ , addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip := ipv4.String()
			if !encountered[ ip ] {
				encountered[ ip ] = true
				ip_addresses = append( ip_addresses , ip )
			}
		}
	}
	return
}

// https://stackoverflow.com/questions/48439363/missing-location-in-call-to-time-in
// thanks bro
func get_location( name string ) ( loc *time.Location ) {
	bs , err := ZoneInfoFS.ReadFile( "zoneinfo/" + name )
	if err != nil { panic( err ) }
	loc , err = time.LoadLocationFromTZData( name , bs )
	if err != nil { panic( err ) }
	return loc
}

func FormatTime( input_time *time.Time ) ( result string ) {
	// location , _ := time.LoadLocation( "America/New_York" )
	// location := get_location( "America/New_York" )
	time_object := input_time.In( location )
	month_name := strings.ToUpper( time_object.Format( "Jan" ) )
	milliseconds := time_object.Format( ".000" )
	date_part := fmt.Sprintf( "%02d%s%d" , time_object.Day() , month_name , time_object.Year() )
	time_part := fmt.Sprintf( "%02d:%02d:%02d%s" , time_object.Hour() , time_object.Minute() , time_object.Second() , milliseconds )
	result = fmt.Sprintf( "%s === %s" , date_part , time_part )
	return
}

func GetFormattedTimeString() ( result string ) {
	// location , _ := time.LoadLocation( "America/New_York" )
	// location := get_location( "America/New_York" )
	time_object := time.Now().In( location )
	month_name := strings.ToUpper( time_object.Format( "Jan" ) )
	milliseconds := time_object.Format( ".000" )
	date_part := fmt.Sprintf( "%02d%s%d" , time_object.Day() , month_name , time_object.Year() )
	time_part := fmt.Sprintf( "%02d:%02d:%02d%s" , time_object.Hour() , time_object.Minute() , time_object.Second() , milliseconds )
	result = fmt.Sprintf( "%s === %s" , date_part , time_part )
	return
}

func GetNowTimeOBJ() ( result time.Time ) {

	result = time.Now().In( location )
	return
}

func GetNowDateString( now *time.Time ) ( result string ) {
	month_name := strings.ToUpper( now.Format( "Jan" ) )
	result = fmt.Sprintf( "%02d%s%d" , now.Day() , month_name , now.Year() )
	return
}

func GetNowTimeString( now *time.Time ) ( result string ) {
	milliseconds := now.Format( ".000" )
	result = fmt.Sprintf( "%02d:%02d:%02d%s" , now.Hour() , now.Minute() , now.Second() , milliseconds )
	return
}

func GetFormattedTimeStringOBJ() ( result_string string , result_time time.Time ) {
	result_time = time.Now().In( location )
	month_name := strings.ToUpper( result_time.Format( "Jan" ) )
	milliseconds := result_time.Format( ".000" )
	date_part := fmt.Sprintf( "%02d%s%d" , result_time.Day() , month_name , result_time.Year() )
	time_part := fmt.Sprintf( "%02d:%02d:%02d%s" , result_time.Hour() , result_time.Minute() , result_time.Second() , milliseconds )
	result_string = fmt.Sprintf( "%s === %s" , date_part , time_part )
	return
}

func IsStringInArray( target string , array []string ) ( bool ) {
	for _ , value := range array {
		if value == target {
			return true
		}
	}
	return false
}

type Slice struct {
	sort.IntSlice
	indexes []int
}
func ( s Slice ) Swap( i , j int ) {
	s.IntSlice.Swap(i, j)
	s.indexes[i], s.indexes[j] = s.indexes[j], s.indexes[i]
}

func NewSlice( n []int ) *Slice {
	s := &Slice{
		IntSlice: sort.IntSlice(n) ,
		indexes: make( []int , len( n ) ) ,
	}
	for i := range s.indexes {
		s.indexes[i] = i
	}
	return s
}

func ReverseInts( input []int ) []int {
	if len(input) == 0 {
		return input
	}
	return append(ReverseInts(input[1:]), input[0])
}

func Sha256Sum( input string ) ( result string ) {
	hash := sha256.Sum256( []byte( input ) )
	result = hex.EncodeToString( hash[ : ] )
	return
}

func CountUniqueViewsInRecords( records []string ) ( result int ) {
	ip_map := map[string]int{}
	for _ , record := range records {
		ip_address := strings.Split( record , " === " )[ 2 ]
		if _ , exists := ip_map[ ip_address ]; exists == false {
			ip_map[ ip_address ] = 1
		}
	}
	result = len( ip_map )
	return
}

func RemoveNonASCII( input string ) ( result string ) {
	for _ , i := range input {
		if i > unicode.MaxASCII { continue }
		result += string( i )
	}
	return
}

const SanitizedStringSizeLimit = 100
func SanitizeInputString( input string ) ( result string ) {
	trimmed := strings.TrimSpace( input )
	if len( trimmed ) > SanitizedStringSizeLimit { trimmed = strings.TrimSpace( trimmed[ 0 : SanitizedStringSizeLimit ] ) }
	result = RemoveNonASCII( trimmed )
	return
}

func ParseFormValueAsInt( context *fiber.Ctx , form_key string ) ( result int ) {
	result = -1
	uploaded := context.FormValue( form_key )
	sanitized := SanitizeInputString( uploaded )
	parsed_int , _ := strconv.Atoi( sanitized )
	result = parsed_int
	return
}

func WriteAdminUserHandOffHTML( server_base_url string ) {
	file , _ := os.OpenFile( "./v1/server/html/admin_user_new_handoff.html" , os.O_RDWR , 0 )
	defer file.Close()
	reader := bufio.NewReader( file )
	line_number := 1
	var lines []string
	for {
		line , err := reader.ReadString( '\n' )
		if err != nil { break }
		if line_number == 48 { line = "\t\t\tconst QR_CODE_BASE_URL = \"" + server_base_url + "\";\n" }
		lines = append( lines , line )
		line_number = line_number + 1
	}
	file.Seek( 0 , 0 )
	file.Truncate( 0 )
	for _ , line := range lines {
		file.WriteString( line )
	}
}

func WriteJS_API( server_base_url string , server_api_key string , local_host_url string ) {
	file, _ := os.OpenFile("./v1/server/cdn/api.js", os.O_RDWR, 0)
	defer file.Close()
	reader := bufio.NewReader(file)
	line_number := 1
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if line_number == 1 {
			line = "const ServerAPIKey = \"" + server_api_key + "\";\n"
		}
		if line_number == 2 {
			line = "const ServerBaseURL = \"" + server_base_url + "\";\n"
		}
		if line_number == 3 {
			line = "const LocalHostURL = \"" + local_host_url + "\";\n"
		}
		lines = append(lines, line)

		if err != nil {
			break
		}
		line_number++
	}
	file.Seek(0, 0)
	file.Truncate(0)
	for _, line := range lines {
		file.WriteString(line)
	}
}

func GenerateNewKeys() {
	fiber_cookie_key := fiber_cookie.GenerateKey()
	bolt_db_key := encryption.GenerateRandomString( 32 )
	server_api_key := encryption.GenerateRandomString( 16 )
	admin_username := encryption.GenerateRandomString( 16 )
	admin_password := encryption.GenerateRandomString( 16 )
	fmt.Println( "Generated New Keys :" )
	fmt.Printf( "\tFiber Cookie Key === %s\n" , fiber_cookie_key )
	fmt.Printf( "\tBolt DB Key === %s\n" , bolt_db_key )
	fmt.Printf( "\tServer API Key === %s\n" , server_api_key )
	fmt.Printf( "\tAdmin Username === %s\n" , admin_username )
	fmt.Printf( "\tAdmin Password === %s\n\n" , admin_password )
}

func PrettyPrint( x_input interface{} ) {
	pretty_json , _ := json.MarshalIndent( x_input , "" , "  " )
	fmt.Println( string( pretty_json ) )
}

func _finger_print_cpu() ( result string ) {
	cpu_info , err := cpu.Info()
	if err != nil || len( cpu_info ) == 0 { return "" }
	info := cpu_info[ 0 ]
	parts := []string{}
	if info.VendorID != "" {
		parts = append( parts , info.VendorID )
	}
	if info.Family != "" {
		parts = append( parts , info.Family )
	}
	if info.Model != "" {
		parts = append( parts , info.Model )
	}
	if info.Cores > 0 {
		parts = append( parts , fmt.Sprintf( "%d" , info.Cores ) )
	}
	if info.ModelName != "" {
		parts = append( parts , info.ModelName )
	}
	return strings.Join( parts , " === " )
}

func _finger_print_mac_address() ( []string ) {
	net_interfaces , _ := net.Interfaces()
	var mac_addresses []string
	for _ , net_interface := range net_interfaces {
		mac_addr := net_interface.HardwareAddr.String()
		if len( mac_addr ) == 0 { continue }
		mac_addresses = append( mac_addresses , mac_addr )
	}
	return mac_addresses
}

type FingerPrintStore struct {
	Sha256 string `json:"sha_256"`
	FingerPrint string `json:"finger_print"`
}
func FingerPrint( config *types.ConfigFile  ) ( result string ) {

	db , _ := bolt_api.Open( config.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	x_cpu_info := _finger_print_cpu()
	x_os := runtime.GOOS
	x_arch := runtime.GOARCH
	x_hostname , _ := os.Hostname()
	x_user , _ := user.Current()
	x_username := x_user.Username
	// x_ip_addresses := GetLocalIPAddresses()
	// x_mac_addresses := _finger_print_mac_address()
	// fmt.Println( x_cpu_info )
	// fmt.Println( x_os )
	// fmt.Println( x_arch )
	// fmt.Println( x_hostname )
	// fmt.Println( x_username )
	// fmt.Println( x_ip_addresses )
	// fmt.Println( x_mac_addresses )
	finger_print_string := fmt.Sprintf( "%s === %s === %s === %s === %s" ,
		x_username ,
		x_os ,
		x_arch ,
		x_hostname ,
		x_cpu_info ,
	)
	finger_print_sha_256 := Sha256Sum( finger_print_string )

	var x_finger_print FingerPrintStore
	db.Update( func( tx *bolt_api.Tx ) error {
		finger_prints_bucket , _ := tx.CreateBucketIfNotExists( []byte( "fingerprints" ) )
		finger_print := finger_prints_bucket.Get( []byte( finger_print_sha_256 ) )
		if finger_print == nil { // Store new fingerprint
			// fmt.Println( "Storing New Finger Print" )
			x_finger_print.Sha256 = finger_print_sha_256
			x_finger_print.FingerPrint = finger_print_string
			x_finger_print_byte_object , _ := json.Marshal( x_finger_print )
			x_finger_print_byte_object_encrypted := encryption.ChaChaEncryptBytes( config.BoltDBEncryptionKey , x_finger_print_byte_object )
			finger_prints_bucket.Put( []byte( finger_print_sha_256 ) , x_finger_print_byte_object_encrypted )
		} else { // Retrieve existing fingerprint
			// fmt.Println( "Retrieving Existing Finger Print" )
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( config.BoltDBEncryptionKey , finger_print )
			json.Unmarshal( decrypted_bucket_value , &x_finger_print )
			if x_finger_print.Sha256 == "" {
				x_finger_print.Sha256 = finger_print_sha_256
				x_finger_print_byte_object , _ := json.Marshal( x_finger_print )
				x_finger_print_byte_object_encrypted := encryption.ChaChaEncryptBytes( config.BoltDBEncryptionKey , x_finger_print_byte_object )
				finger_prints_bucket.Put( []byte( finger_print_sha_256 ) , x_finger_print_byte_object_encrypted )
			}
		}
		return nil
	})
	fmt.Println( x_finger_print )
	result = x_finger_print.Sha256
	return
}

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