package redis

import (
	"fmt"
	"context"
	"sync"
	"reflect"
	"strconv"
	redis_manager "github.com/0187773933/RedisManagerUtils/manager"
	redis_lib "github.com/go-redis/redis/v8"
	// types "github.com/0187773933/MastersCloset/v1/types"
)

func GetConnection( host string , port string , db int , password string ) ( redis_client redis_manager.Manager ) {
	redis_client.Connect( fmt.Sprintf( "%s:%s" , host , port ) , db , password )
	// TODO ===
	// https://stackoverflow.com/questions/49910285/redis-dial-tcp-redis-address-connect-connection-refused#49910392
	// redis := &redis.Pool{
	// 	MaxActive: idleConnections,
	// 	MaxIdle:   idleConnections,
	// 	Wait:      true,
	// 	Dial: func() (redis.Conn, error) {
	// 		c, err := redis.Dial("tcp", address, options...)
	// 		for retries := 0; err != nil && retries < 5; retries++ {
	// 			time.Sleep((50 << retries) * time.Millisecond)
	// 			c, err = redis.Dial("tcp", address, options...)
	// 		}
	// 		return c, err
	// 	},
	// }
	return
}

func SetAdd( redis redis_manager.Manager , set_key string , value string ) {
	var ctx = context.Background()
	_ , set_add_error := redis.Redis.SAdd( ctx , set_key , value ).Result()
	if set_add_error != nil { fmt.Println( set_add_error ); }
	return
}

func GetList( redis redis_manager.Manager , list_key string ) ( result []string ) {
	var ctx = context.Background()
	list_result , list_get_error := redis.Redis.LRange( ctx , list_key , 0 , -1 ).Result()
	if list_get_error != nil { fmt.Println( list_get_error ); }
	result = list_result
	return
}

func Get( redis redis_manager.Manager , key string , wg *sync.WaitGroup , result_channel chan string ) {
	defer wg.Done()
	var ctx = context.Background()
	fmt.Println( key )
	string_result , string_result_error := redis.Redis.Get( ctx , key ).Result()
	if string_result_error != nil { fmt.Println( string_result_error ); }
	// result = string_result
	if len( string_result ) > 0 {
		fmt.Println( "1 - here" , reflect.TypeOf( string_result ) , len( string_result ) , string_result )
		result_channel <- string_result
	} else {
		result_channel <- "empty"
	}
	return
}

func GetMulti( redis redis_manager.Manager , uuids []string ) ( result []string ) {
	var ctx = context.Background()
	pipe := redis.Redis.Pipeline()
	var values []*redis_lib.StringCmd
	for _ , uuid := range uuids {
		// this assumes all commands in multi are "GETS"
		values = append( values , pipe.Get( ctx , uuid ) )
		// but we don't care about values , we need :
			// SMEMBERS for ANALYTICS.{UUID}.IPS
			// LRANGE for ANALYTICS.{UUID}.RECORDS
			// GET for ANALYTICS.{UUID}.TOTAL
	}
	pipe.Exec( ctx )
	pipe.Close()
	for _ , i := range values {
		result = append( result , i.Val() )
	}
	return
}

type CustomMulti struct {
	Totals []int
	IPS [][]string
	Records [][]string
	Names []string
}
func Multi( redis redis_manager.Manager , uuids []string ) ( results CustomMulti ) {
	var ctx = context.Background()
	pipe := redis.Redis.Pipeline()

	// so each multi requires a currated set of typed receivers that can later be "unwrapped"
	// aka .Result() can be called once the pipe.Exec() is finished
	// you have to know the type each pipe.Function() will return before hand , so you know what type of bins to create
	// TODO = Add Name if Set

	// could make just an Interface{}
	// then it doesn't matter the type
	// but still here , we have to have a custom Redis->Multi() order
	var ones []*redis_lib.StringCmd
	var twos []*redis_lib.StringSliceCmd
	var threes []*redis_lib.StringSliceCmd
	var names []*redis_lib.StringCmd
	for _ , uuid := range uuids {
		one := pipe.Get( ctx , fmt.Sprintf( "ANALYTICS.%s.TOTAL" , uuid ) )
		two := pipe.SMembers( ctx , fmt.Sprintf( "ANALYTICS.%s.IPS" , uuid ) )
		three := pipe.LRange( ctx , fmt.Sprintf( "ANALYTICS.%s.RECORDS" , uuid ) , 0 , -1 )
		name := pipe.Get( ctx , fmt.Sprintf( "ANALYTICS.%s.NAME" , uuid ) )
		ones = append( ones , one )
		twos = append( twos , two )
		threes = append( threes , three )
		names = append( names , name )
	}
	pipe.Exec( ctx )
	pipe.Close()

	// unwrap the results
	for _ , e := range ones {
		value , _ := e.Result()
		int_value , _ := strconv.Atoi( value )
		results.Totals = append( results.Totals , int_value )
	}
	// fmt.Println( "original , original totals" , results.Totals )
	for _ , e := range twos {
		value , _ := e.Result()
		results.IPS = append( results.IPS , value )
	}
	for _ , e := range threes {
		value , _ := e.Result()
		results.Records = append( results.Records , value )
	}
	for _ , e := range names {
		value , _ := e.Result()
		results.Names = append( results.Names , value )
	}
	return
}