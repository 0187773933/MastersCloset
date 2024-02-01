package adminroutes

import (
	"fmt"
	"time"
	// "strconv"
	// "math/rand"
	// "strings"
	json "encoding/json"
	// uuid "github.com/satori/go.uuid"
	// short_uuid "github.com/lithammer/shortuuid/v4"
	// rid "github.com/solutionroute/rid"
	// aaa "github.com/nii236/adjectiveadjectiveanimal"
	fiber "github.com/gofiber/fiber/v2"
	// types "github.com/0187773933/MastersCloset/v1/types"
	// bolt "github.com/0187773933/MastersCloset/v1/bolt"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	// bleve "github.com/blevesearch/bleve/v2"
	// utils "github.com/0187773933/MastersCloset/v1/utils"
	// log "github.com/0187773933/MastersCloset/v1/log"
)

// lots of opinions on how to do this
// haversine , euclidean , levenshtein , or just identical
func _user_similar_by_name( sent_user *user.User , compared_user *user.User ) ( result bool ) {
	result = false
	if sent_user.Identity.FirstName == "" { return }
	if sent_user.Identity.LastName == "" { return }
	if compared_user.Identity.FirstName == "" { return }
	if compared_user.Identity.LastName == "" { return }
	if sent_user.Identity.FirstName == compared_user.Identity.FirstName {
		// if sent_user.Identity.MiddleName == compared_user.Identity.MiddleName {
			if sent_user.Identity.LastName == compared_user.Identity.LastName {
				result = true
			}
		// }
	}
	return
}

func _user_similar_by_email( sent_user *user.User , compared_user *user.User ) ( result bool ) {
	result = false
	if sent_user.EmailAddress == "" { return }
	if compared_user.EmailAddress == "" { return }
	if sent_user.EmailAddress == compared_user.EmailAddress {
		result = true
	}
	return
}

func _user_similar_by_phone( sent_user *user.User , compared_user *user.User ) ( result bool ) {
	result = false
	if sent_user.PhoneNumber == "" { return }
	if compared_user.PhoneNumber == "" { return }
	if sent_user.PhoneNumber == compared_user.PhoneNumber {
		result = true
	}
	return
}

func _user_similar_by_address( sent_user *user.User , compared_user *user.User ) ( result bool ) {
	result = false
	if sent_user.Identity.Address.StreetNumber == "" { return }
	if sent_user.Identity.Address.StreetName == "" { return }
	if compared_user.Identity.Address.StreetNumber == "" { return }
	if compared_user.Identity.Address.StreetName == "" { return }
	if sent_user.Identity.Address.StreetNumber == compared_user.Identity.Address.StreetNumber {
		if sent_user.Identity.Address.StreetName == compared_user.Identity.Address.StreetName {
			result = true
		}
	}
	return
}

func _user_similar_by_birthday( sent_user *user.User , compared_user *user.User ) ( result bool ) {
	result = false
	if sent_user.Identity.DateOfBirth.Year == 0 { return }
	if sent_user.Identity.DateOfBirth.Month == "" { return }
	if sent_user.Identity.DateOfBirth.Day == 0 { return }
	if compared_user.Identity.DateOfBirth.Year == 0 { return }
	if compared_user.Identity.DateOfBirth.Month == "" { return }
	if compared_user.Identity.DateOfBirth.Day == 0 { return }
	if sent_user.Identity.DateOfBirth.Year == compared_user.Identity.DateOfBirth.Year {
		if sent_user.Identity.DateOfBirth.Month == compared_user.Identity.DateOfBirth.Month {
			if sent_user.Identity.DateOfBirth.Day == compared_user.Identity.DateOfBirth.Day {
				result = true
				return
			}
		}
	}
	return
}

func _user_similar_by_barcode( sent_user *user.User , compared_user *user.User ) ( result bool ) {
	result = false
	for _ , sent_barcode := range sent_user.Barcodes {
		if sent_barcode == "" { continue }
		for _ , compared_barcode := range compared_user.Barcodes {
			if compared_barcode == "" { continue }
			if sent_barcode == compared_barcode {
				result = true
				return
			}
		}
	}
	return
}

type UserSimilarReport struct {
	IsSimilar bool `json:"is_similar"`
	Name bool `json:"name"`
	Email bool `json:"email"`
	Phone bool `json:"phone"`
	Address bool `json:"address"`
	Birthday bool `json:"birthday"`
	Barcode bool `json:"barcode"`
	User user.User `json:"user"`
}
func _user_is_similar( sent_user *user.User , compared_user *user.User ) ( result UserSimilarReport ) {
	result.Name = _user_similar_by_name( sent_user , compared_user )
	if result.Name == true { result.IsSimilar = true }
	result.Email = _user_similar_by_email( sent_user , compared_user )
	if result.Email == true { result.IsSimilar = true }
	result.Phone = _user_similar_by_phone( sent_user , compared_user )
	if result.Phone == true { result.IsSimilar = true }
	result.Address = _user_similar_by_address( sent_user , compared_user )
	if result.Address == true { result.IsSimilar = true }
	result.Birthday = _user_similar_by_birthday( sent_user , compared_user )
	if result.Birthday == true { result.IsSimilar = true }
	result.Barcode = _user_similar_by_barcode( sent_user , compared_user )
	if result.Barcode == true { result.IsSimilar = true }
	return
}

func HandleUserSimilar( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	var sent_user user.User
	context_body := context.Body()
	fmt.Println( string( context_body ) )
	json.Unmarshal( context_body , &sent_user )
	var similar_user_reports []UserSimilarReport
	db , _ := bolt_api.Open( GlobalConfig.BoltDBPath , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()
	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		bucket.ForEach( func( uuid , value []byte ) error {
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( GlobalConfig.BoltDBEncryptionKey , value )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			similarity_report := _user_is_similar( &sent_user , &viewed_user )
			if similarity_report.IsSimilar == false { return nil }
			similarity_report.User = viewed_user
			similar_user_reports = append( similar_user_reports , similarity_report )
			return nil
		})
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/similar" ,
		"sent_user": sent_user ,
		"similar_user_reports": similar_user_reports ,
	})
}