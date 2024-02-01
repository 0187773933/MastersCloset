package adminroutes

import (
	// "fmt"
	// "time"
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
	// bolt_api "github.com/boltdb/bolt"
	user "github.com/0187773933/MastersCloset/v1/user"
	// encryption "github.com/0187773933/MastersCloset/v1/encryption"
	// bolt "github.com/boltdb/bolt"
	// bleve "github.com/blevesearch/bleve/v2"
	// utils "github.com/0187773933/MastersCloset/v1/utils"
	log "github.com/0187773933/MastersCloset/v1/log"
)

func HandleUserSimilar( context *fiber.Ctx ) ( error ) {
	if validate_admin_session( context ) == false { return serve_failed_attempt( context ) }
	var viewed_user user.User
	json.Unmarshal( context.Body() , &viewed_user )
	log.Debug( "Viewed User === %v\n" , viewed_user )
	return context.JSON( fiber.Map{
		"route": "/admin/user/new" ,
		"sent_user": viewed_user ,
		"similar_users": []user.User{} ,
	})
}