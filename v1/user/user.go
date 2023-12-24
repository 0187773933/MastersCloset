package user

import (
	"fmt"
	// "reflect"
	// "strconv"
	"strings"
	"time"
	json "encoding/json"
	// bolt "github.com/0187773933/MastersCloset/v1/bolt"
	bolt "github.com/boltdb/bolt"
	bleve "github.com/blevesearch/bleve/v2"
	uuid "github.com/satori/go.uuid"
	encrypt "github.com/0187773933/MastersCloset/v1/encryption"
	types "github.com/0187773933/MastersCloset/v1/types"
	log "github.com/0187773933/MastersCloset/v1/log"
	printer "github.com/0187773933/MastersCloset/v1/printer"
)

type CheckIn struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	ULID string `json:"ULID"`
	Date string `json:"date"`
	Time string `json:"time"`
	Type string `json:"type"`
	Result bool `json:"result"`
	TimeRemaining int `json:"time_remaining"`
	PrintJob printer.PrintJob `json:"print_job"`
}

// type FailedCheckIn struct {
// 	Date string `json:"date"`
// 	Time string `json:"time"`
// 	Type string `json:"type"`
// 	DaysRemaining int `json:"remaining_days"`
// }

type BalanceItem struct {
	Available int `json:"available"`
	Limit int `json:"limit"`
	Used int `json:"used"`
}

type GeneralClothes struct {
	Total int `json:"total"`
	Available int `json:"available"`
	Tops BalanceItem `json:"tops"`
	Bottoms BalanceItem `json:"bottoms"`
	Dresses BalanceItem `json:"dresses"`
}

type Balance struct {
	General GeneralClothes `json:"general"`
	Shoes BalanceItem `json:"shoes"`
	Seasonals BalanceItem `json:"seasonals"`
	Accessories BalanceItem `json:"accessories"`
}

type DateOfBirth struct {
	Month string `json:"month"`
	Day int `json:"day"`
	Year int `json:"year"`
}

type Address struct {
	StreetNumber string `json:"street_number"`
	StreetName string `json:"street_name"`
	AddressTwo string `json:"address_two"`
	City string `json:"city"`
	State string `json:"state"`
	ZipCode string `json:"zipcode"`
}

type Person struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Address Address`json:"address"`
	DateOfBirth DateOfBirth `json:"date_of_birth"`
	Age int `json:"age"`
	Sex string `json:"sex"`
	Height string `json:"height"`
	EyeColor string `json:"eye_color"`
	Spouse bool `json:"spouse"`
}

type User struct {
	Config *types.ConfigFile `json:"-"`
	Verified bool `json:"verified"`
	Username string `json:"username"`
	NameString string `json:"name_string"`
	UUID string `json:"uuid"`
	ULID string `json:"ulid"`
	Barcodes []string `json:"barcodes"`
	EmailAddress string `json:"email_address"`
	PhoneNumber string `json:"phone_number"`
	Identity Person `json:"identity"`
	AuthorizedAliases []Person `json:"authorized_aliases"`
	FamilySize int `json:"family_size"`
	FamilyMembers []Person `json:"family_members"`
	CreatedDate string `json:"created_date"`
	CreatedTime string `json:"created_time"`
	CheckIns []CheckIn `json:"check_ins"`
	FailedCheckIns []CheckIn `json:"failed_check_ins"`
	Balance Balance `json:"balance"`
	TimeRemaining int `json:"time_remaining"`
	AllowedToCheckIn bool `json:"allowed_to_checkin"`
	Spanish bool `json:"spanish"`
}

type GetUserResult struct {
	Username string `json:"username"`
	UUID string `json:"uuid"`
	LastCheckIn CheckIn `json:"last_check_in"`
}

func New( username string , config *types.ConfigFile ) ( new_user User ) {
	now := time.Now()
	new_user_uuid := uuid.NewV4().String()
	new_user.Username = username
	new_user.Verified = false
	new_user.UUID = new_user_uuid
	new_user.Config = config
	new_user.FamilySize = 1
	new_user.CreatedDate = now.Format( "02JAN2006" )
	new_user.CreatedTime = now.Format( "15:04:05.000" )
	new_user_byte_object , _ := json.Marshal( new_user )
	new_user_byte_object_encrypted := encrypt.ChaChaEncryptBytes( config.BoltDBEncryptionKey , new_user_byte_object )
	db , _ := bolt.Open( config.BoltDBPath , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
	db_result := db.Update( func( tx *bolt.Tx ) error {
		users_bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )
		users_bucket.Put( []byte( new_user_uuid ) , new_user_byte_object_encrypted )
		usernames_bucket , _ := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
		// ideally : bcrypted first and last name search --> base64 salsabox username --> uuid --> user
		// but we have bleve search
		// would have to encrypt decrypt it each time
		// holomorphic search ?
		usernames_bucket.Put( []byte( username ) , []byte( new_user.UUID ) )

		if new_user.ULID != "" {
			fmt.Println( "ulid was present" )
			ulid_uuid_bucket , _ := tx.CreateBucketIfNotExists( []byte( "ulid-uuid" ) )
			ulid_uuid_bucket.Put( []byte( new_user.ULID ) , []byte( new_user.UUID ) )
		}

		return nil
	})
	db.Close()
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }
	return
}

func ( u *User ) UpdateSelfFromDB() {
	db , _ := bolt.Open( u.Config.BoltDBPath , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	db_result := db.View( func( tx *bolt.Tx ) error {
		users_bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )
		x_user := users_bucket.Get( []byte( u.UUID ) )
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( u.Config.BoltDBEncryptionKey , x_user )
		json.Unmarshal( decrypted_bucket_value , &u )
		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }
}

func ( u *User ) Save() {
	db , _ := bolt.Open( u.Config.BoltDBPath , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	var existing_user *User
	u.FormatUsername()
	db_result := db.Update( func( tx *bolt.Tx ) error {

		// this was originally the only thing in here
		users_bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )

		// but we added stuff below now on every save

		// Grab existing version of user to see if we need to make any adjacent db changes
		existing_user_value := users_bucket.Get( []byte( u.UUID ) )
		if existing_user_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( u.Config.BoltDBEncryptionKey , existing_user_value )
		json.Unmarshal( decrypted_bucket_value , &existing_user )

		// such as the usernames bucket
		usernames_bucket , _ := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
		if existing_user.Username != u.Username {
			// fmt.Println( "we have to update the username for search and stuff" )
			// fmt.Println( "existing username: " + existing_user.Username )
			// fmt.Println( "new: " + u.Username )
			usernames_bucket.Delete( []byte( existing_user.Username ) )
			search_index , _ := bleve.Open( u.Config.BleveSearchPath )
			defer search_index.Close()
			edited_search_item := types.SearchItem{
				UUID: u.UUID ,
				Name: u.NameString ,
			}
			search_index.Index( u.UUID , edited_search_item )
		}
		usernames_bucket.Put( []byte( u.Username ) , []byte( u.UUID ) )

		// and the barcode bucket
		barcodes_bucket , _ := tx.CreateBucketIfNotExists( []byte( "barcodes" ) )
		for i := 0; i < len( u.Barcodes ); i++ {
			barcodes_bucket.Put( []byte( u.Barcodes[ i ] ) , []byte( u.UUID ) )
			// TODO , handle what happens if we remove a barcode from a user
			// Not really that big of a problem , since this just updates the barcode for the right uuid anyway
		}

		byte_object , _ := json.Marshal( u )
		byte_object_encrypted := encrypt.ChaChaEncryptBytes( u.Config.BoltDBEncryptionKey , byte_object )
		users_bucket.Put( []byte( u.UUID ) , byte_object_encrypted )

		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }
}

func ( u *User ) Delete() {
	// byte_object , _ := json.Marshal( u )
	// byte_object_encrypted := encrypt.ChaChaEncryptBytes( u.Config.BoltDBEncryptionKey , byte_object )
	// db , _ := bolt.Open( u.Config.BoltDBPath , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
	// defer db.Close()
	// db_result := db.Update( func( tx *bolt.Tx ) error {
	// 	users_bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )
	// 	users_bucket.Put( []byte( u.UUID ) , byte_object_encrypted )
	// 	return nil
	// })
	// if db_result != nil { panic( "couldn't write to bolt db ??" ) }
}

func ( u *User ) RefillBalance() {
	u.GetFamilySize()
	u.Balance.General.Total = ( u.Config.Balance.General.Total * u.FamilySize )
	u.Balance.General.Available = ( u.Config.Balance.General.Total * u.FamilySize )
	u.Balance.General.Tops.Limit = ( u.Config.Balance.General.Tops * u.FamilySize )
	u.Balance.General.Tops.Available = ( u.Config.Balance.General.Tops * u.FamilySize )
	u.Balance.General.Bottoms.Limit = ( u.Config.Balance.General.Bottoms * u.FamilySize )
	u.Balance.General.Bottoms.Available = ( u.Config.Balance.General.Bottoms * u.FamilySize )
	u.Balance.General.Dresses.Limit = ( u.Config.Balance.General.Dresses * u.FamilySize )
	u.Balance.General.Dresses.Available = ( u.Config.Balance.General.Dresses * u.FamilySize )
	u.Balance.Shoes.Limit = ( u.Config.Balance.Shoes * u.FamilySize )
	u.Balance.Shoes.Available = ( u.Config.Balance.Shoes * u.FamilySize )
	u.Balance.Seasonals.Limit = ( u.Config.Balance.Seasonals * u.FamilySize )
	u.Balance.Seasonals.Available = ( u.Config.Balance.Seasonals * u.FamilySize )
	u.Balance.Accessories.Limit = ( u.Config.Balance.Accessories * u.FamilySize )
	u.Balance.Accessories.Available = ( u.Config.Balance.Accessories * u.FamilySize )
	u.Save()
	return
}

func ( u *User ) GetFamilySize() ( result int ) {
	result = ( len( u.FamilyMembers ) + 1 )
	if ( u.FamilySize != result ) {
		u.FamilySize = result
		u.Save()
	}
	return
}

func (u *User) FormatUsername() {
	var username_parts []string
	if u.Identity.FirstName != "" {
		u.Identity.FirstName = strings.Title( strings.ToLower( strings.TrimSpace( u.Identity.FirstName ) ) )
		username_parts = append( username_parts , u.Identity.FirstName )
	}
	if u.Identity.MiddleName != "" {
		u.Identity.MiddleName = strings.Title( strings.ToLower( strings.TrimSpace( u.Identity.MiddleName ) ) )
		username_parts = append( username_parts , u.Identity.MiddleName )
	}
	if u.Identity.LastName != "" {
		u.Identity.LastName = strings.Title( strings.ToLower( strings.TrimSpace( u.Identity.LastName ) ) )
		username_parts = append( username_parts , u.Identity.LastName )
	}
	if len(username_parts) > 0 {
		u.Username = strings.Join( username_parts , "-" )
		u.NameString = strings.Join( username_parts , " " )
	}
}

func ( u *User ) CheckInTest() ( check_in CheckIn ) {

	// 1.) prelim
	time_remaining := -1
	// db , _ := bolt.Open( u.Config.BoltDBPath , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
	// defer db.Close()

	// 2.) Test if Check-In is possible
	now := time.Now()
	now_time_zone := now.Location()
	check_in.Date = now.Format( "02Jan2006" )
	check_in.Time = now.Format( "15:04:05.000" )

	var lockout_message string
	if len( u.CheckIns ) < 1 {
		check_in.Result = true
		time_remaining = 0
	} else {
		// user has checked in before , need to compare last check-in date to now
		// only comparing the dates , not the times
		last_check_in := u.CheckIns[ len( u.CheckIns ) - 1 ]
		last_check_in_date , _ := time.ParseInLocation( "02Jan2006" , last_check_in.Date , now_time_zone )
		// log.Debug( "Now ===" , now )
		// log.Debug( "Last ===" , last_check_in_date )

		cool_off_hours := ( 24 * u.Config.CheckInCoolOffDays )
		// log.Debug( "Cooloff Hours ===" , cool_off_hours )
		cool_off_duration , _ := time.ParseDuration( fmt.Sprintf( "%dh" , cool_off_hours ) )
		// log.Debug( "Cooloff Duration ===" , cool_off_duration )

		check_in_date_difference := now.Sub( last_check_in_date )
		// log.Debug( "Difference ===" , check_in_date_difference )

		// Negative Values Mean The User Has Waited Long Enough
		// Positive Values Mean the User Still has to wait
		time_remaining_duration := ( cool_off_duration - check_in_date_difference )
		// log.Debug( "Time Remaining ===" , time_remaining_duration )

		if time_remaining_duration < 0 {
			// "the user waited long enough before checking in again"
			check_in.Result = true
			time_remaining = 0
		} else {

			days_remaining := int( time_remaining_duration.Hours() / 24 )
			time_remaining_string := time_remaining_duration.String()

			// lockout_message := fmt.Sprintf( "The user did NOT wait long enough before checking in again , has to wait : %d days , or %s" , days_remaining , time_remaining_string )
			lockout_message = fmt.Sprintf( "=== has to wait : %d days , or %s" , days_remaining , time_remaining_string )

			check_in.Result = false
			time_remaining = int( time_remaining_duration.Milliseconds() )
		}
	}
	u.TimeRemaining = time_remaining
	u.AllowedToCheckIn = check_in.Result
	// log.PrintlnConsole( u.NameString , "===" , "Allowed To Check In ===" , check_in.Result , lockout_message )
	log.PrintlnConsole( u.UUID , "===" , "Allowed To Check In ===" , check_in.Result , lockout_message )
	check_in.Date = strings.ToUpper( check_in.Date )
	check_in.TimeRemaining = time_remaining
	return
}

func ( u *User ) CheckIn() ( check_in CheckIn ) {
	check_in = u.CheckInTest()
	if ( u.AllowedToCheckIn == false ) {
		log.Debug( "User timed out" )
		u.FailedCheckIns = append( u.FailedCheckIns , check_in )
		u.Save()
		return
	}
	if check_in.Result == true {
		// "the user waited long enough before checking in again"
		check_in.Type = "new"
		u.CheckIns = append( u.CheckIns , check_in )
	} else {
		log.Debug( "time remaining ===" , check_in.TimeRemaining )
	}
	u.Save()
	return
}

func ( u *User ) CheckInForce() ( check_in CheckIn ) {
	check_in = u.CheckInTest()
	check_in.Type = "forced"
	check_in.Result = true
	check_in.TimeRemaining = 0
	u.CheckIns = append( u.CheckIns , check_in )
	u.Save()
	return
}

func ( u *User ) AddBarcode( barcode string ) {
	u.Barcodes = append( u.Barcodes , barcode )
	u.Save()
	return
}

func FormatUsername( x_user *User ) {
	var username_parts []string
	if x_user.Identity.FirstName != "" {
		x_user.Identity.FirstName = strings.Title( strings.ToLower( strings.TrimSpace( x_user.Identity.FirstName ) ) )
		username_parts = append( username_parts , x_user.Identity.FirstName )
	}
	if x_user.Identity.MiddleName != "" {
		x_user.Identity.MiddleName = strings.Title( strings.ToLower( strings.TrimSpace( x_user.Identity.MiddleName ) ) )
		username_parts = append( username_parts , x_user.Identity.MiddleName )
	}
	if x_user.Identity.LastName != "" {
		x_user.Identity.LastName = strings.Title( strings.ToLower( strings.TrimSpace( x_user.Identity.LastName ) ) )
		username_parts = append( username_parts , x_user.Identity.LastName )
	}
	if len( username_parts ) > 0 {
		x_user.Username = strings.Join( username_parts , "-" )
		x_user.NameString = strings.Join( username_parts , " " )
	}
}

func UserNameExists( username string , db *bolt.DB ) ( result bool , uuid string ) {
	result = false
	db.Update( func( tx *bolt.Tx ) error {
		bucket , tx_error := tx.CreateBucketIfNotExists( []byte( "usernames" ) )
		if tx_error != nil { log.Debug( tx_error ); return nil }
		bucket_value := bucket.Get( []byte( username ) )
		if bucket_value == nil { return nil }
		result = true
		uuid = string( bucket_value )
		return nil
	})
	return
}

// renaming
func GetByUUID( user_uuid string , db *bolt.DB , encryption_key string ) ( viewed_user User ) {
	db.View( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		// log.Debug( string( decrypted_bucket_value ) )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	return
}

func GetViaUUID( user_uuid string , config *types.ConfigFile ) ( viewed_user User ) {
	db , _ := bolt.Open( config.BoltDBPath , 0600 , &bolt.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	db.View( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( config.BoltDBEncryptionKey , bucket_value )
		// log.Debug( string( decrypted_bucket_value ) )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	viewed_user.Config = config
	return
}

func RefillBalance( user_uuid string , db *bolt.DB , encryption_key string , balance_config types.BalanceConfig , family_size int ) ( new_balance Balance ) {
	var viewed_user User
	db.Update( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )

		viewed_user.Balance.General.Total = ( balance_config.General.Total * family_size )
		viewed_user.Balance.General.Available = ( balance_config.General.Total * family_size )
		viewed_user.Balance.General.Tops.Limit = ( balance_config.General.Tops * family_size )
		viewed_user.Balance.General.Tops.Available = ( balance_config.General.Tops * family_size )
		viewed_user.Balance.General.Bottoms.Limit = ( balance_config.General.Bottoms * family_size )
		viewed_user.Balance.General.Bottoms.Available = ( balance_config.General.Bottoms * family_size )
		viewed_user.Balance.General.Dresses.Limit = ( balance_config.General.Dresses * family_size )
		viewed_user.Balance.General.Dresses.Available = ( balance_config.General.Dresses * family_size )
		viewed_user.Balance.Shoes.Limit = ( balance_config.Shoes * family_size )
		viewed_user.Balance.Shoes.Available = ( balance_config.Shoes * family_size )
		viewed_user.Balance.Seasonals.Limit = ( balance_config.Seasonals * family_size )
		viewed_user.Balance.Seasonals.Available = ( balance_config.Seasonals * family_size )
		viewed_user.Balance.Accessories.Limit = ( balance_config.Accessories * family_size )
		viewed_user.Balance.Accessories.Available = ( balance_config.Accessories * family_size )

		viewed_user_byte_object , _ := json.Marshal( viewed_user )
		viewed_user_byte_object_encrypted := encrypt.ChaChaEncryptBytes( encryption_key , viewed_user_byte_object )
		bucket.Put( []byte( user_uuid ) , viewed_user_byte_object_encrypted )
		return nil
	})
	new_balance = viewed_user.Balance
	return
}


// non-volitle? / passive checkin
// just sees if its possible , or if the user is currently timed-out
// 8e1bb28c-8868-448f-a07e-f0d270b4bbee === should be able to check-in
// d1e22369-6777-4eff-bf6a-0bf46a343a72
func CheckInTest( user_uuid string , db *bolt.DB , encryption_key string , cool_off_days int ) ( result bool , time_remaining int , balance Balance , name_string string , family_size int ) {
	result = false
	time_remaining = -1
	// 1.) grab the user from the db
	var viewed_user User
	db.View( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	if viewed_user.UUID == "" { log.Debug( "user UUID doesn't exist" ); result = false; return }

	// 2.) Test if Check-In is possible
	var new_check_in CheckIn
	now := time.Now()
	now_time_zone := now.Location()
	new_check_in.Date = now.Format( "02Jan2006" )
	new_check_in.Time = now.Format( "15:04:05.000" )

	if len( viewed_user.CheckIns ) < 1 {
		result = true
		time_remaining = 0
	} else {
		// user has checked in before , need to compare last check-in date to now
		// only comparing the dates , not the times
		last_check_in := viewed_user.CheckIns[ len( viewed_user.CheckIns ) - 1 ]
		last_check_in_date , _ := time.ParseInLocation( "02Jan2006" , last_check_in.Date , now_time_zone )
		log.Debug( "Now ===" , now )
		log.Debug( "Last ===" , last_check_in_date )

		cool_off_hours := ( 24 * cool_off_days )
		log.Debug( "Cooloff Hours ===" , cool_off_hours )
		cool_off_duration , _ := time.ParseDuration( fmt.Sprintf( "%dh" , cool_off_hours ) )
		log.Debug( "Cooloff Duration ===" , cool_off_duration )

		check_in_date_difference := now.Sub( last_check_in_date )
		log.Debug( "Difference ===" , check_in_date_difference )

		// Negative Values Mean The User Has Waited Long Enough
		// Positive Values Mean the User Still has to wait
		time_remaining_duration := ( cool_off_duration - check_in_date_difference )
		log.Debug( "Time Remaining ===" , time_remaining_duration )

		if time_remaining_duration < 0 {
			// "the user waited long enough before checking in again"
			result = true
			time_remaining = 0
		} else {

			days_remaining := int( time_remaining_duration.Hours() / 24 )
			time_remaining_string := time_remaining_duration.String()
			fmt.Printf( "the user did NOT wait long enough before checking in again , has to wait : %d days , or %s\n" , days_remaining , time_remaining_string )

			result = false
			time_remaining = int( time_remaining_duration.Milliseconds() )
		}
	}
	balance = viewed_user.Balance
	name_string = viewed_user.NameString
	family_size = 1
	if viewed_user.FamilySize > 0 {
		family_size = viewed_user.FamilySize
	}

	return
}

func CheckInUser( user_uuid string , db *bolt.DB , encryption_key string , cool_off_days int ) ( result bool , time_remaining int ) {
	result = false
	time_remaining = -1
	var viewed_user User
	db.View( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encrypt.ChaChaDecryptBytes( encryption_key , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		return nil
	})
	if viewed_user.UUID == "" { log.Debug( "user UUID doesn't exist" ); result = false; return }
	var new_check_in CheckIn
	now := time.Now()
	now_time_zone := now.Location()
	new_check_in.Date = now.Format( "02Jan2006" )
	new_check_in.Time = now.Format( "15:04:05.000" )
	if len( viewed_user.CheckIns ) < 1 {
		new_check_in.Type = "first"
		new_check_in.Date = strings.ToUpper( new_check_in.Date )
		viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )
		result = true
		time_remaining = 0
	} else {
		// user has checked in before , need to compare last check-in date to now
		// only comparing the dates , not the times
		last_check_in := viewed_user.CheckIns[ len( viewed_user.CheckIns ) - 1 ]
		last_check_in_date , _ := time.ParseInLocation( "02Jan2006" , last_check_in.Date , now_time_zone )
		log.Debug( "Now/New ===" , now )
		log.Debug( "Last ===" , last_check_in_date )

		cool_off_hours := ( 24 * cool_off_days )
		log.Debug( "Cooloff Hours ===" , cool_off_hours )
		cool_off_duration , _ := time.ParseDuration( fmt.Sprintf( "%dh" , cool_off_hours ) )
		log.Debug( "Cooloff Duration ===" , cool_off_duration )

		check_in_date_difference := now.Sub( last_check_in_date )
		log.Debug( "Difference ===" , check_in_date_difference )

		// Negative Values Mean The User Has Waited Long Enough
		// Positive Values Mean the User Still has to wait
		time_remaining_duration := ( cool_off_duration - check_in_date_difference )
		log.Debug( "Time Remaining ===" , time_remaining_duration )

		if time_remaining_duration < 0 {
			// "the user waited long enough before checking in again"
			new_check_in.Type = "new"
			viewed_user.CheckIns = append( viewed_user.CheckIns , new_check_in )
			result = true
			time_remaining = 0
		} else {

			days_remaining := int( time_remaining_duration.Hours() / 24 )
			time_remaining_string := time_remaining_duration.String()
			fmt.Printf( "the user did NOT wait long enough before checking in again , has to wait : %d days , or %s\n" , days_remaining , time_remaining_string )

			time_remaining = int( time_remaining_duration.Milliseconds() )
			new_check_in.TimeRemaining = time_remaining
			new_check_in.Result = false
			viewed_user.FailedCheckIns = append( viewed_user.FailedCheckIns , new_check_in )

			result = false
		}
	}
	viewed_user_byte_object , _ := json.Marshal( viewed_user )
	viewed_user_byte_object_encrypted := encrypt.ChaChaEncryptBytes( encryption_key , viewed_user_byte_object )
	db_result := db.Update( func( tx *bolt.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.Put( []byte( user_uuid ) , viewed_user_byte_object_encrypted )
		return nil
	})
	if db_result != nil { panic( "couldn't write to bolt db ??" ) }
	return
}