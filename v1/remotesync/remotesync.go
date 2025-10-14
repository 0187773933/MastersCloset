package remotesync

import (
	"bytes"
	"context"
	// "encoding/binary"
	// "errors"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	bolt "github.com/boltdb/bolt"
	// encrypt "github.com/0187773933/MastersCloset/v1/encryption"
	types "github.com/0187773933/MastersCloset/v1/types"
)

var INTERVAL = 12 * time.Second
var UPLOAD_BUCKET_NAME = "remote-upload"
var DOWNLOAD_BUCKET_NAME = "remote-download"

type RemoteSync struct {
	DB *bolt.DB `json:"-"`
	CTX context.Context `json:"-"`
	CONFIG *types.ConfigFile `jsong:"-"`
}

func New( db *bolt.DB , ctx context.Context , config *types.ConfigFile ) ( result RemoteSync ) {
	result.DB = db
	result.CTX = ctx
	result.CONFIG = config
	result.DB.Update( func( tx *bolt.Tx ) error {
		_ , err := tx.CreateBucketIfNotExists( []byte( UPLOAD_BUCKET_NAME ) )
		if err == nil { return err }
		_ , err = tx.CreateBucketIfNotExists( []byte( DOWNLOAD_BUCKET_NAME ) )
		return err
	})
	return
}

func ( rs *RemoteSync ) Start() {
	go func() {
		timer := time.NewTimer( INTERVAL )
		defer timer.Stop()

		http_client := &http.Client{ Timeout: 10 * time.Second }

		for {
			select {
				case <-rs.CTX.Done():
					fmt.Println( "simulated context expired ?" )
					return
				case <-timer.C:
					fmt.Println( "uploading all modified users" )
					rs.DB.Update( func( tx *bolt.Tx ) error {
						b := tx.Bucket( []byte( UPLOAD_BUCKET_NAME ) )
						b.ForEach( func( k , v []byte ) error {
							fmt.Printf( "A %s is %s.\n" , k , v )
							upload_result := rs.Upload( http_client , &k , &v )
							fmt.Println( upload_result )
							return nil
						})
						return nil
					})
					timer.Reset( INTERVAL )
			}
		}
	}()
}

func ( rs *RemoteSync ) Upload( client *http.Client , uuid *[]byte , u_bytes *[]byte ) ( result bool ) {
	result = false
	req , err := http.NewRequest( "POST" , rs.CONFIG.RemoteHostUrl , bytes.NewReader( *u_bytes ) )
	if err != nil {
		fmt.Println( err )
		result = false
		return
	}
	req.Header.Set( "Content-Type" , "application/json" )
	resp , err := client.Do( req )
	if err != nil {
		fmt.Println( err )
		result = false
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet , _ := io.ReadAll( io.LimitReader( resp.Body , 2048 ) )
		x := fmt.Errorf( "http %d: %s", resp.StatusCode , string( snippet ) )
		fmt.Println( x )
		result = false
		return
	}
	var json_response map[string]any
	if err := json.NewDecoder( io.LimitReader( resp.Body , 1<<20 ) ).Decode( &json_response ); err != nil {
		x := fmt.Errorf( "decode response: %w" , err )
		fmt.Println( x )
		result = false
		return
	}
	remote_result , _ := json_response[ "result" ].( bool )
	if !remote_result {
		x := fmt.Errorf( "remote result=false" )
		fmt.Println( x )
		result = false
		return
	}
	result = true
	return
}
