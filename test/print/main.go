package main

import (
	"fmt"
	"time"
	// "strings"
	json "encoding/json"
	bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	utils "github.com/0187773933/MastersCloset/v1/utils"
	printer "github.com/0187773933/MastersCloset/v1/printer"
	types "github.com/0187773933/MastersCloset/v1/types"
)

type PrinterConfig struct {
	PageWidth float64 `json:"page_width"`
	PageHeight float64 `json:"page_height"`
	FontName string `json:"font_name"`
	PrinterName string `json:"printer_name"`
	LogoFilePath string `json:"logo_file_path"`
}

func main() {

	bolt_db_path := "/Users/morpheous/WORKSPACE/GO/MastersClosetTracker/mct.db"
	bolt_db_key := ""
	db , _ := bolt_api.Open( bolt_db_path , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	user_uuid := ""
	var viewed_user user.User

	db.View( func( tx *bolt_api.Tx ) error {
		users_bucket := tx.Bucket( []byte( "users" ) )
		bucket_value := users_bucket.Get( []byte( user_uuid ) )
		if bucket_value == nil { return nil }
		decrypted_bucket_value := encryption.ChaChaDecryptBytes( bolt_db_key , bucket_value )
		json.Unmarshal( decrypted_bucket_value , &viewed_user )
		// user.FormatUsername( &viewed_user )
		fmt.Println( viewed_user.NameString )
		return nil
	})

	printer_config := &types.PrinterConfig{
		PageWidth: 4 ,
		PageHeight: 6 ,
		FontName: "ComicNeue" ,
		PrinterName: "_4BARCODE_4B_2054N" ,
		LogoFilePath: "./v1/server/cdn/logo.png" ,
	}

	print_job := printer.PrintJob{
		FamilySize: 2 ,
		TotalClothingItems: 42 ,
		Shoes: 7 ,
		ShoesLimit: 1 ,
		Accessories: 14 ,
		AccessoriesLimit: 2 ,
		Seasonal: 7 ,
		SeasonalLimit: 1 ,
		FamilyName: viewed_user.Identity.LastName ,
		BarcodeNumber: "1234" ,
		Spanish: true ,
	}
	fmt.Println( "Printing Ticket :" , print_job )
	utils.PrettyPrint( print_job )
	printer.PrintTicket( *printer_config , print_job )

}
