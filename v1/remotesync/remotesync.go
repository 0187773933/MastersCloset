package remotesync

import (
	// "bytes"
	"context"
	// "encoding/binary"
	// "errors"
	"fmt"
	// "io"
	// "net/http"
	"time"
	bolt "github.com/boltdb/bolt"
	// encrypt "github.com/0187773933/MastersCloset/v1/encryption"
	types "github.com/0187773933/MastersCloset/v1/types"
)

var INTERVAL = 12 * time.Second
var BUCKET_NAME = "remote-save"

type RemoteSync struct {
	DB *bolt.DB `json:"-"`
	CTX context.Context `json:"-"`
	Config *types.ConfigFile `jsong:"-"`
}

func New( db *bolt.DB , ctx context.Context , config *types.ConfigFile ) ( result RemoteSync ) {
	result.DB = db
	result.CTX = ctx
	result.Config = config
	return
}

func ( rs *RemoteSync ) EnsureRemoteSaveBucket() ( err error ) {
	return rs.DB.Update( func( tx *bolt.Tx ) error {
		_ , err := tx.CreateBucketIfNotExists( []byte( BUCKET_NAME ) )
		return err
	})
}

func ( rs *RemoteSync ) Start() {
	go func() {
		timer := time.NewTimer( INTERVAL )
		defer timer.Stop()

		// client := &http.Client{ Timeout: 10 * time.Second }

		for {
			select {
				case <-rs.CTX.Done():
					fmt.Println( "simulated context expired ?" )
					return
				case <-timer.C:
					fmt.Println( "simulated que drain" )
					// Drain the queue until empty.
					// for {
					// 	payload , ok , err := popOne( db )
					// 	if err != nil {
					// 		// DB issue; try again next cycle
					// 		break
					// 	}
					// 	if !ok {
					// 		break
					// 	}
					// 	// Do HTTP outside the txn.
					// 	if err := doPost(client, postURL, payload); err != nil {
					// 		// Requeue on failure (best-effort) and back off slightly
					// 		_ = EnqueueRemoteSave(db, payload)
					// 		time.Sleep(500 * time.Millisecond)
					// 	}
					// }
					// // Schedule next run AFTER finishing this one.
					timer.Reset( INTERVAL )
			}
		}
	}()
}

// // Pop the oldest item (atomic within one Update).
// // Returns (payload, ok, err). ok=false means queue empty.
// func popOne(db *bolt.DB) ([]byte, bool, error) {
// 	var vcopy []byte
// 	err := db.Update(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte(RemoteSaveBucket))
// 		if b == nil {
// 			return nil
// 		}
// 		c := b.Cursor()
// 		k, v := c.First()
// 		if k == nil {
// 			return nil
// 		}
// 		vcopy = append([]byte(nil), v...)
// 		return c.Delete()
// 	})
// 	if err != nil {
// 		return nil, false, err
// 	}
// 	if vcopy == nil {
// 		return nil, false, nil
// 	}
// 	return vcopy, true, nil
// }

// func doPost(client *http.Client, url string, payload []byte) error {
// 	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
// 	if err != nil {
// 		return err
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	io.Copy(io.Discard, resp.Body)
// 	resp.Body.Close()
// 	if resp.StatusCode >= 300 {
// 		return fmt.Errorf("http %d", resp.StatusCode)
// 	}
// 	return nil
// }
