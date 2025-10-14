package server

import (
	// "os"
	"fmt"
	"bytes"
	"strings"
	json "encoding/json"
	net_url "net/url"
	fiber "github.com/gofiber/fiber/v2"
	bolt_api "github.com/boltdb/bolt"
	bleve "github.com/blevesearch/bleve/v2"
	user "github.com/0187773933/MastersCloset/v1/user"
	encryption "github.com/0187773933/MastersCloset/v1/encryption"
	log "github.com/0187773933/MastersCloset/v1/log"
)

func ( s *Server ) UserSearch( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }

	username := context.Params( "username" )
	escaped_username , _ := net_url.QueryUnescape( username )
	formated_username := strings.Replace( escaped_username , " " , "-" , -1 )
	formated_username_bytes := []byte( formated_username )
	found_uuid := "not found"
	s.DB.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "usernames" ) )
		bucket.ForEach( func( k , v []byte ) error {
			if bytes.Equal( k , formated_username_bytes ) == false { return nil }
			found_uuid = string( v )
			return nil
		})
		return nil
	})
	log.Debug( fmt.Sprintf( "Searched : %s || Result === %s\n" , formated_username , found_uuid ) )
	return context.JSON( fiber.Map{
		"route": "/admin/user/search/username/:username" ,
		"result": found_uuid ,
	})
}

func ( s *Server ) UserSearchFuzzy( context *fiber.Ctx ) ( error ) {
	if s.ValidateAdminSession( context ) == false { return s.ServeFailedAttempt( context ) }

	username := context.Params( "username" )
	escaped_username , _ := net_url.QueryUnescape( username )
	escaped_username = strings.ToLower( escaped_username ) // have to fix db first

	search_index , _ := bleve.Open( s.Config.BleveSearchPath )
	defer search_index.Close()

	// Only works for single words
	// query := bleve.NewFuzzyQuery( escaped_username )
	// query.Fuzziness = 2
	// // query.Fuzziness = 1
	// search_request := bleve.NewSearchRequest( query )
	// search_results , _ := search_index.Search( search_request )

	words := strings.Fields( escaped_username )
	boolean_query := bleve.NewBooleanQuery()
	for _ , word := range words {
		q := bleve.NewFuzzyQuery( word )
		q.Fuzziness = 1
		q.SetField( "Name" )
		boolean_query.AddMust( q )
	}
	search_request := bleve.NewSearchRequest( boolean_query )
	search_results , _ := search_index.Search( search_request )

	var search_results_users []user.User
	s.DB.View( func( tx *bolt_api.Tx ) error {
		bucket := tx.Bucket( []byte( "users" ) )
		for _ , hit := range search_results.Hits {
			x_user := bucket.Get( []byte( hit.ID ) )
			if x_user == nil { continue } // so this is needed because we didn't delete search indexes when deleting a user
			var viewed_user user.User
			decrypted_bucket_value := encryption.ChaChaDecryptBytes( s.Config.BoltDBEncryptionKey , x_user )
			json.Unmarshal( decrypted_bucket_value , &viewed_user )
			search_results_users = append( search_results_users , viewed_user )
		}
		return nil
	})
	return context.JSON( fiber.Map{
		"route": "/admin/user/search/username/fuzzy/:username" ,
		"result": search_results_users ,
	})
}