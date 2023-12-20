package adminroutes

import (
	"fmt"
	"time"
	"strings"
	"strconv"
	json "encoding/json"
	bolt_api "github.com/boltdb/bolt"
	fiber "github.com/gofiber/fiber/v2"
	utils "github.com/0187773933/MastersCloset/v1/utils"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	user "github.com/0187773933/MastersCloset/v1/user"
	printer "github.com/0187773933/MastersCloset/v1/printer"
	log "github.com/0187773933/MastersCloset/v1/log"
	ulid "github.com/oklog/ulid/v2"
)

type CheckInBalanceForm struct {
	TopsAvailable int `json:"balance_tops_available"`
	TopsLimit int `json:"balance_tops_limit"`
	TopsUsed int `json:"balance_tops_used"`
	BottomsAvailable int `json:"balance_bottoms_available"`
	BottomsLimit int `json:"balance_bottoms_limit"`
	PantsLimit int `json:"balance_pants_limit"`
	BottomsUsed int `json:"balance_bottoms_used"`
	DressesAvailable int `json:"balance_dresses_available"`
	DressesLimit int `json:"balance_dresses_limit"`
	DressesUsed int `json:"balance_dresses_used"`
	ShoesAvailable int `json:"balance_shoes_available"`
	ShoesLimit int `json:"balance_shoes_limit"`
	ShoesUsed int `json:"balance_shoes_used"`
	SeasonalsAvailable int `json:"balance_seasonals_available"`
	SeasonalsLimit int `json:"balance_seasonals_limit"`
	SeasonalsUsed int `json:"balance_seasonals_used"`
	AccessoriesAvailable int `json:"balance_accessories_available"`
	AccessoriesLimit int `json:"balance_accessories_limit"`
	AccessoriesUsed int `json:"balance_accessories_used"`
	ShoppingFor int `json:"shopping_for"`
	Boys int `json:"boys"`
	Girls int `json:"girls"`
	Men int `json:"men"`
	Women int `json:"women"`
}

// Refill If Empty , Subtract Check-In Ticket
func _ries( user_item *int , ticket_value int , limit int ) {
	if *user_item < 1 {
		*user_item = limit
	}
	*user_item = ( *user_item - ticket_value )
}

// Add To Total
func _att( user_item *int , amount int ) {
	// fmt.Println( "setting" , *user_item , amount )
	*user_item = ( *user_item + amount )
}

// We changed this to a POST Form , so now we have to parse it
func UserCheckIn( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }

	var balance_form CheckInBalanceForm
	json.Unmarshal( []byte( context.Body() ), &balance_form )
	fmt.Println( "hello ???" , "UserCheckIn()" )
	fmt.Printf( "%+v\n" , balance_form )

	// 1.) Prep
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	// 2.) Grab the User
	uploaded_uuid := context.Params( "uuid" )
	x_uuid := utils.SanitizeInputString( uploaded_uuid )
	var viewed_user user.User
	viewed_user.Config = GlobalConfig
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( x_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})

	// 3.) Create a New Check In
	var new_check_in user.CheckIn
	now := time.Now()
	// now_time_zone := now.Location()
	new_check_in.Date = now.Format( "02Jan2006" )
	new_check_in.Time = now.Format( "15:04:05.000" )
	new_check_in.Type = "forced"
	new_check_in.Date = strings.ToUpper( new_check_in.Date )
	// new_check_in.ShoppingFor = balance_form.ShoppingFor
	// viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )

	// 4.) Update the Balance
	_att( &viewed_user.Balance.General.Tops.Used , balance_form.TopsAvailable )
	_ries( &viewed_user.Balance.General.Tops.Available , balance_form.TopsAvailable , ( GlobalConfig.Balance.General.Tops * balance_form.ShoppingFor ) )

	_att( &viewed_user.Balance.General.Bottoms.Used , balance_form.BottomsAvailable )
	_ries( &viewed_user.Balance.General.Bottoms.Available , balance_form.BottomsAvailable , ( GlobalConfig.Balance.General.Bottoms * balance_form.ShoppingFor ) )

	_att( &viewed_user.Balance.General.Dresses.Used , balance_form.DressesAvailable )
	_ries( &viewed_user.Balance.General.Dresses.Available , balance_form.DressesAvailable , ( GlobalConfig.Balance.General.Dresses * balance_form.ShoppingFor ) )

	_att( &viewed_user.Balance.Shoes.Used , balance_form.ShoesAvailable )
	_ries( &viewed_user.Balance.Shoes.Available , balance_form.ShoesAvailable , ( GlobalConfig.Balance.Shoes * balance_form.ShoppingFor ) )

	_att( &viewed_user.Balance.Seasonals.Used , balance_form.SeasonalsAvailable )
	_ries( &viewed_user.Balance.Seasonals.Available , balance_form.SeasonalsAvailable , ( GlobalConfig.Balance.Seasonals * balance_form.ShoppingFor ) )

	_att( &viewed_user.Balance.Accessories.Used , balance_form.AccessoriesAvailable )
	_ries( &viewed_user.Balance.Accessories.Available , balance_form.AccessoriesAvailable , ( GlobalConfig.Balance.Accessories * balance_form.ShoppingFor ) )

	// fmt.Println( "Checking In With Balance :" )
	// fmt.Printf( "%+v\n" , viewed_user.Balance )

	// 5.) Print Ticket
	// TODO : clarify calculation ????
	// total_clothing_items := ( balance_form.TopsAvailable + balance_form.BottomsAvailable + balance_form.DressesAvailable )
	// total_clothing_items := ( balance_form.TopsAvailable + balance_form.ShoesAvailable + balance_form.SeasonalsAvailable + balance_form.AccessoriesAvailable )
	// total_clothing_items := ( balance_form.TopsAvailable )
	total_clothing_items := ( balance_form.ShoppingFor * 6 )
	// family_size := viewed_user.FamilySize
	// if family_size < 1 { family_size = 1 } // this is what happens when you don't just use sql
	barcode_number := ""
	if len( viewed_user.Barcodes ) > 0 && len( viewed_user.Barcodes[ 0 ] ) > 1 {
		barcode_number = viewed_user.Barcodes[ 0 ]
	} else {
		// Create a high number barcode
		// barcode_number = viewed_user.AddVirtualBarcode()
		db.Update( func( tx *bolt_api.Tx ) error {
			misc_bucket , _ := tx.CreateBucketIfNotExists( []byte( "misc" ) )
			vb_index_bucket_value := misc_bucket.Get( []byte( "virtual-barcode-index" ) )
			// fmt.Println( "vb_index_bucket_value" , vb_index_bucket_value )
			vb_index := 9999999
			if vb_index_bucket_value != nil {
				vb_index , _ = strconv.Atoi( string( vb_index_bucket_value ) )
			}
			vb_index = vb_index + 1

			barcode_number = strconv.Itoa( vb_index )
			log.PrintlnConsole( "Adding Virtual Barcode :" , barcode_number )
			misc_bucket.Put( []byte( "virtual-barcode-index" ) , []byte( barcode_number ) )
			viewed_user.Barcodes = append( viewed_user.Barcodes , barcode_number )

			barcodes_bucket , _ := tx.CreateBucketIfNotExists( []byte( "barcodes" ) )
			barcodes_bucket.Put( []byte( barcode_number ) , []byte( viewed_user.UUID ) )

			return nil
		})
	}

	family_name := viewed_user.NameString
	// if len( family_name ) > 20 ? // TODO : Find max length of family string

	// TODO : add people list ???
	// change ticket design ????
	print_job := printer.PrintJob{
		FamilySize: balance_form.ShoppingFor ,
		TotalClothingItems: total_clothing_items ,
		Shoes: balance_form.ShoesAvailable ,
		ShoesLimit: GlobalConfig.Balance.Shoes ,
		PantsLimit: GlobalConfig.Balance.General.Bottoms ,
		Accessories: balance_form.AccessoriesAvailable ,
		AccessoriesLimit: GlobalConfig.Balance.Accessories ,
		Seasonal: balance_form.SeasonalsAvailable ,
		SeasonalLimit: GlobalConfig.Balance.Seasonals ,
		FamilyName: family_name ,
		BarcodeNumber: barcode_number ,
		Spanish: viewed_user.Spanish ,
		Boys: balance_form.Boys ,
		Girls: balance_form.Girls ,
		Men: balance_form.Men ,
		Women: balance_form.Women ,
	}

	new_check_in.UUID = viewed_user.UUID
	new_check_in.Name = viewed_user.NameString
	new_check_in.ULID = ulid.Make().String()
	new_check_in.PrintJob = print_job
	viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )

	// if len( GlobalConfig.LocalHostUrl ) > 3 {
	// 	log.PrintlnConsole( "Printing Ticket :" , print_job )
	// 	utils.PrettyPrint( print_job )
	// 	printer.PrintTicket( GlobalConfig.Printer , print_job )
	// }

	// 6.) Re-Save the User
	viewed_user_byte_object , _ := json.Marshal( viewed_user )
	viewed_user_byte_object_encrypted := encryption.ChaChaEncryptBytes( GlobalConfig.BoltDBEncryptionKey , viewed_user_byte_object )
	db_result := db.Update( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.Put( []byte( x_uuid ) , viewed_user_byte_object_encrypted )
		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }

	// 7.) Return Result
	return context.JSON( fiber.Map{
		"route": "/admin/user/checkin/:uuid" ,
		"result": true ,
		"check_in": new_check_in ,
	})
}


func UserCheckInTest( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	x_user_uuid := context.Params( "uuid" )
	x_user := user.GetViaUUID( x_user_uuid , GlobalConfig )
	check_in_test := x_user.CheckInTest()
	return context.JSON( fiber.Map{
		"route": "/admin/user/checkin/test/:uuid" ,
		"result": check_in_test ,
		"user": x_user ,
		"balance_config": GlobalConfig.Balance ,
	})
}