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
	HTTPClient *http.Client `json:"-"`
}

func New( db *bolt.DB , ctx context.Context , config *types.ConfigFile ) ( result RemoteSync ) {
	result.DB = db
	result.CTX = ctx
	result.CONFIG = config
	result.HTTPClient = &http.Client{ Timeout: 10 * time.Second }
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

		for {
			select {
				case <-rs.CTX.Done():
					fmt.Println( "simulated context expired ?" )
					return
				case <-timer.C:
					// fmt.Println( "uploading all modified users" )
					rs.DB.Update( func( tx *bolt.Tx ) error {

						m_b , _ := tx.CreateBucketIfNotExists( []byte( "MISC" ) )
						last_download_sequence_id := string( m_b.Get( []byte( "remote-last-download-sequence" ) ) )
						last_upload_sequence_id := string( m_b.Get( []byte( "remote-last-upload-sequence" ) ) )

						users_bucket , _ := tx.CreateBucketIfNotExists( []byte( "users" ) )

						// download any changes
						remote_changes := rs.DownloadChangedUsersList( last_download_sequence_id )
						total_remote_changed := len( remote_changes )
						for i , remote_change := range remote_changes {
							fmt.Printf( "Downloading [ %d ] of %d changed\n" , ( i + 1 ) , total_remote_changed )
							downloaded_user := rs.DownloadUser( remote_change.UUID )

							users_bucket.Put( []byte( downloaded_user.UUID ) , downloaded_user.UserBytes )
							m_b.Put( []byte( "remote-last-download-sequence" ) , []byte( remote_change.ID ) )
							fmt.Printf( "\t%s === %s\n" , remote_change.ID , downloaded_user.UUID )
						}

						// TODO :: check if there are any changed uuids before downloading
						b , _ := tx.CreateBucketIfNotExists( []byte( UPLOAD_BUCKET_NAME ) )
						total_changed := b.Stats().KeyN
						if total_changed == 0 {
							// fmt.Println( "No Users To Upload" )
							return nil
						}

						// Upload all local changes
						i := 1
						b.ForEach( func( k , v []byte ) error {
							// fmt.Printf( "A %s is %s.\n" , k , v )
							upload_result := rs.Upload( &k , &v )
							fmt.Printf( "Uploading [ %d ] of %d Edited Users ... Success === %t\n" , i , total_changed , upload_result.Result )
							if upload_result.Result == true {
								b.Delete( k )
								m_b.Put( []byte( "remote-last-upload-sequence" ) , []byte( upload_result.Sequence ) )
								last_upload_sequence_id = upload_result.Sequence
							}
							i += 1
							return nil
						})

						fmt.Println( last_upload_sequence_id )

						return nil
					})
					timer.Reset( INTERVAL )
			}
		}
	}()
}

type UploadResult struct {
	Result bool `json:"result"`
	Sequence string `json:"sequence"`
}
func ( rs *RemoteSync ) Upload( uuid *[]byte , u_bytes *[]byte ) ( result UploadResult ) {
	req , err := http.NewRequest( "POST" , rs.CONFIG.RemoteHostUrl + "/import" , bytes.NewReader( *u_bytes ) )
	if err != nil {
		fmt.Println( err )
		return
	}
	req.Header.Set( "Content-Type" , "application/json" )
	req.Header.Set( fmt.Sprintf( "%s-CLIENT-ID" , rs.CONFIG.RemoteHostHeaderPrefix ) , rs.CONFIG.RemoteHostClientID )
	req.Header.Set( fmt.Sprintf( "%s-UUID" , rs.CONFIG.RemoteHostHeaderPrefix ) , string( *uuid ) )
	req.Header.Set( fmt.Sprintf( "%s-API-KEY" , rs.CONFIG.RemoteHostHeaderPrefix ) , rs.CONFIG.RemoteHostAPIKey )
	resp , err := rs.HTTPClient.Do( req )
	if err != nil {
		fmt.Println( err )
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet , _ := io.ReadAll( io.LimitReader( resp.Body , 2048 ) )
		x := fmt.Errorf( "http %d: %s", resp.StatusCode , string( snippet ) )
		fmt.Println( x )
		return
	}
	body_bytes , _ := io.ReadAll( io.LimitReader( resp.Body , 1<<20 ) )
	if err := json.Unmarshal( body_bytes , &result ); err != nil {
		x := fmt.Errorf( "decode response: %w" , err )
		fmt.Println( x )
		return
	}
	fmt.Println( result )
	return
}

type ChangedUser struct {
	ID string `json:"id"`
	UUID string `json:"uuid"`
}
type ChangedUserListResult struct {
	Changes []ChangedUser `json:"changes"`
}
func ( rs *RemoteSync ) DownloadChangedUsersList( since_sequence_id string ) ( result []ChangedUser ) {
	req , err := http.NewRequest( "GET" , rs.CONFIG.RemoteHostUrl + "/changed" , nil )
	if err != nil {
		fmt.Println( err )
		return
	}
	req.Header.Set( "Content-Type" , "application/json" )
	req.Header.Set( fmt.Sprintf( "%s-CLIENT-ID" , rs.CONFIG.RemoteHostHeaderPrefix ) , rs.CONFIG.RemoteHostClientID )
	req.Header.Set( fmt.Sprintf( "%s-API-KEY" , rs.CONFIG.RemoteHostHeaderPrefix ) , rs.CONFIG.RemoteHostAPIKey )
	req.Header.Set( fmt.Sprintf( "%s-SEQUENCE-ID" , rs.CONFIG.RemoteHostHeaderPrefix ) , since_sequence_id )
	resp , err := rs.HTTPClient.Do( req )
	if err != nil {
		fmt.Println( err )
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet , _ := io.ReadAll( io.LimitReader( resp.Body , 2048 ) )
		fmt.Printf( "HTTP %d: %s\n" , resp.StatusCode , snippet )
		return
	}
	body_bytes , _ := io.ReadAll( io.LimitReader( resp.Body , 1<<20 ) )
	var result_struct ChangedUserListResult
	if err = json.Unmarshal( body_bytes , &result_struct ); err != nil {
		fmt.Println( "decode error:" , err )
		return
	}
	result = result_struct.Changes
	return
}

type DownloadedUser struct {
	UUID string `json:"uuid"`
	UserBytes []byte `json:"user_bytes"`
}
func ( rs *RemoteSync ) DownloadUser( uuid string ) ( result DownloadedUser ) {
	req , err := http.NewRequest( "GET" , rs.CONFIG.RemoteHostUrl + "/download" , nil )
	if err != nil {
		fmt.Println( err )
		return
	}
	req.Header.Set( "Content-Type" , "application/json" )
	req.Header.Set( fmt.Sprintf( "%s-UUID" , rs.CONFIG.RemoteHostHeaderPrefix ) , uuid )
	req.Header.Set( fmt.Sprintf( "%s-API-KEY" , rs.CONFIG.RemoteHostHeaderPrefix ) , rs.CONFIG.RemoteHostAPIKey )
	resp , err := rs.HTTPClient.Do( req )
	if err != nil {
		fmt.Println( err )
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet , _ := io.ReadAll( io.LimitReader( resp.Body , 2048 ) )
		fmt.Printf( "HTTP %d: %s\n" , resp.StatusCode , snippet )
		return
	}
	body_bytes , _ := io.ReadAll( io.LimitReader( resp.Body , 1<<20 ) )
	if err = json.Unmarshal( body_bytes , &result ); err != nil {
		fmt.Println( "decode error:" , err )
		return
	}
	return
}


