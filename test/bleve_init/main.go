package main

import (
	"fmt"
	"time"
	"strings"
	bolt_api "github.com/boltdb/bolt"
	bleve "github.com/blevesearch/bleve/v2"
)

type SearchItem struct {
	UUID string
	Name string
}

func main() {

	bolt_db_path := "/Users/morpheous/WORKSPACE/GO/MastersCloset/mct.db"
	db , _ := bolt_api.Open( bolt_db_path , 0600 , &bolt_api.Options{ Timeout: ( 3 * time.Second ) } )
	defer db.Close()

	index_path := "/Users/morpheous/WORKSPACE/GO/MastersCloset/mct.bleve"
	search_mapping := bleve.NewIndexMapping()
	search , _ := bleve.New( index_path , search_mapping )
	defer search.Close()

	db.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "usernames" ) )
		bucket.ForEach( func( username , uuid []byte ) error {
			if username == nil { return nil }
			if uuid == nil { return nil }
			name := strings.ReplaceAll( string( username ) , "-" , " " )
			name_lower := strings.ToLower( name )
			fmt.Printf( "%s === %s\n" , name_lower , uuid )
			message := SearchItem{
				UUID: string( uuid ) ,
				Name: name_lower ,
			}
			search.Index( message.UUID , message )
			return nil
		})
		return nil
	})
}
