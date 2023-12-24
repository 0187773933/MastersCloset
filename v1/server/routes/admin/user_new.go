package adminroutes

import (
	"fmt"
	"time"
	"strconv"
	"math/rand"
	// "strings"
	json "encoding/json"
	uuid "github.com/satori/go.uuid"
	// short_uuid "github.com/lithammer/shortuuid/v4"
	// rid "github.com/solutionroute/rid"
	// aaa "github.com/nii236/adjectiveadjectiveanimal"
	fiber "github.com/gofiber/fiber/v2"
	types "github.com/0187773933/MastersCloset/v1/types"
	// bolt "github.com/0187773933/MastersCloset/v1/bolt"
	// bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	// encryption "github.com/0187773933/MastersCloset/v1/encryption"
	bolt "github.com/boltdb/bolt"
	bleve "github.com/blevesearch/bleve/v2"
	utils "github.com/0187773933/MastersCloset/v1/utils"
	log "github.com/0187773933/MastersCloset/v1/log"
)

func ProcessNewUserForm( context *fiber.Ctx ) ( new_user user.User ) {

	uploaded_first_name := context.FormValue( "user_first_name" )
	uploaded_last_name := context.FormValue( "user_last_name" )
	uploaded_user_middle_name := context.FormValue( "user_middle_name" )
	uploaded_user_email := context.FormValue( "user_email" )
	uploaded_phone_number := context.FormValue( "user_phone_number" )
	uploaded_user_street_number := context.FormValue( "user_street_number" )
	uploaded_user_street_name := context.FormValue( "user_street_name" )
	uploaded_user_address_two := context.FormValue( "user_address_two" )
	uploaded_user_city := context.FormValue( "user_city" )
	uploaded_user_state := context.FormValue( "user_state" )
	uploaded_user_zip_code := context.FormValue( "user_zip_code" )
	uploaded_user_birth_day := context.FormValue( "user_birth_day" )
	uploaded_user_birth_month := context.FormValue( "user_birth_month" )
	uploaded_user_birth_year := context.FormValue( "user_birth_year" )
	uploaded_user_family_size := context.FormValue( "user_family_size" )
	uploaded_total_barcodes := context.FormValue( "total_barcodes" )

	new_user.EmailAddress = utils.SanitizeInputString( uploaded_user_email )
	new_user.PhoneNumber = utils.SanitizeInputString( uploaded_phone_number )
	new_user.Identity.FirstName = utils.SanitizeInputString( uploaded_first_name )
	new_user.Identity.MiddleName = utils.SanitizeInputString( uploaded_user_middle_name )
	new_user.Identity.LastName = utils.SanitizeInputString( uploaded_last_name )
	new_user.Identity.Address.StreetNumber = utils.SanitizeInputString( uploaded_user_street_number )
	new_user.Identity.Address.StreetName = utils.SanitizeInputString( uploaded_user_street_name )
	new_user.Identity.Address.AddressTwo = utils.SanitizeInputString( uploaded_user_address_two )
	new_user.Identity.Address.City = utils.SanitizeInputString( uploaded_user_city )
	new_user.Identity.Address.State = utils.SanitizeInputString( uploaded_user_state )
	new_user.Identity.Address.ZipCode = utils.SanitizeInputString( uploaded_user_zip_code )

	sanitized_birth_day := utils.SanitizeInputString( uploaded_user_birth_day )
	sanitized_birth_day_int , _ := strconv.Atoi( sanitized_birth_day )
	new_user.Identity.DateOfBirth.Day = sanitized_birth_day_int
	new_user.Identity.DateOfBirth.Month = utils.SanitizeInputString( uploaded_user_birth_month )
	sanitized_birth_year := utils.SanitizeInputString( uploaded_user_birth_year )
	sanitized_birth_year_int , _ := strconv.Atoi( sanitized_birth_year )
	new_user.Identity.DateOfBirth.Year = sanitized_birth_year_int

	sanitized_total_barcodes := utils.SanitizeInputString( uploaded_total_barcodes )
	sanitized_total_barcodes_int , _ := strconv.Atoi( sanitized_total_barcodes )
	if sanitized_total_barcodes_int > 0 {
		for i := 0; i < sanitized_total_barcodes_int; i++ {
			uploaded_barcode := context.FormValue( fmt.Sprintf( "user_barcode_%d" , ( i + 1 ) ) )
			sanitized_barcode := utils.SanitizeInputString( uploaded_barcode )
			new_user.Barcodes = append( new_user.Barcodes , sanitized_barcode )
		}
	}

	sanitized_family_size := utils.SanitizeInputString( uploaded_user_family_size )
	sanitized_family_size_int , _ := strconv.Atoi( sanitized_family_size )
	new_user.FamilySize = sanitized_family_size_int
	if sanitized_family_size_int > 0 {
		for i := 0; i < sanitized_family_size_int; i++ {
			uploaded_family_member_age := context.FormValue( fmt.Sprintf( "user_family_member_%d_age" , ( i + 1 ) ) )
			sanitized_family_member_age := utils.SanitizeInputString( uploaded_family_member_age )
			family_member_age_int , _ := strconv.Atoi( sanitized_family_member_age )
			var family_member user.Person
			family_member.Age = family_member_age_int
			new_user.FamilyMembers = append( new_user.FamilyMembers , family_member )
			// fmt.Printf( "Adding Family Member - %d - Age = %d\n" , ( i + 1 ) , family_member_age_int )
		}
	}

	user.FormatUsername( &new_user )

	new_user.UUID = uuid.NewV4().String()

	now := time.Now()
	new_user.CreatedDate = now.Format( "02JAN2006" )
	new_user.CreatedTime = now.Format( "15:04:05.000" )

	return
}

func HandleNewUserJoin( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	var viewed_user user.User
	json.Unmarshal( context.Body() , &viewed_user )

	viewed_user.Config = GlobalConfig

	// Treat this as a Temp User
	if viewed_user.Identity.FirstName == "" && viewed_user.Identity.MiddleName == "" && viewed_user.Identity.LastName == "" {
		// log.PrintlnConsole( "this was a temp user" )
		viewed_user.Identity.FirstName = "Temp"
		// temp_id := aaa.Generate( 1 , &aaa.Options{} )
		// viewed_user.Identity.MiddleName = strings.Title( temp_id[ 0 ] )
		// viewed_user.Identity.LastName = strings.Title( temp_id[ 1 ] )

		rand.Seed( time.Now().UnixNano() )
		// temp_id := aaa.Generate( 1 , &aaa.Options{} )
		// f_name := fmt.Printf( "%08x" , rand.Uint32() )
		viewed_user.Identity.MiddleName = fmt.Sprintf( "%v%v%v%v%v%v" , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) )
		viewed_user.Identity.LastName = fmt.Sprintf( "%v%v%v%v%v%v" , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) , rand.Intn( 9 ) )
	}

	viewed_user.FormatUsername()

	new_user := user.New( viewed_user.Username , GlobalConfig )

	log.Println( new_user )

	viewed_user.UUID = new_user.UUID
	viewed_user.CreatedDate = new_user.CreatedDate
	viewed_user.CreatedTime = new_user.CreatedTime
	viewed_user.Save()

	if viewed_user.ULID != "" {
		fmt.Println( "ulid was present" )
		db , _ := bolt.Open( GlobalConfig.BoltDBPath , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
		defer db.Close()
		db.Update( func( tx *bolt.Tx ) error {
			ulid_uuid_bucket , _ := tx.CreateBucketIfNotExists( []byte( "ulid-uuid" ) )
			ulid_uuid_bucket.Put( []byte( viewed_user.ULID ) , []byte( viewed_user.UUID ) )
			return nil
		})
	}

	// add to search index
	search_index , _ := bleve.Open( GlobalConfig.BleveSearchPath )
	defer search_index.Close()
	search_item := types.SearchItem{
		UUID: new_user.UUID ,
		Name: viewed_user.NameString ,
	}
	log.Printf( "Updating Search Index with : %s\n" , viewed_user.NameString )
	search_index.Index( new_user.UUID , search_item )

	log.PrintlnConsole( "Created New User :" , viewed_user.NameString )

	// viewed_user.Save();
	return context.JSON( fiber.Map{
		"route": "/admin/user/new" ,
		"result": viewed_user ,
	})
}