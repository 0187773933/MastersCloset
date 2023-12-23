const uuid_v4_regex = /^[0-9A-F]{8}-[0-9A-F]{4}-[4][0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$/i
function is_uuid( str ) { return uuid_v4_regex.test( str ); }
const barcode_regex = /^\d+$/;
function is_barcode( str ) { return barcode_regex.test( str ); }
function sleep( ms ) { return new Promise( resolve => setTimeout( resolve , ms ) ); }

const USER_PROTO = `package user;
syntax = "proto3";

message Address {
	string street_number = 1;
	string street_name = 2;
	string address_two = 3;
	string city = 4;
	string state = 5;
	string zipcode = 8;
}

message DOB {
	int32 day = 1;
	string month = 2;
	int32 year = 3;
}

message Identity {
	string first_name = 1;
	string middle_name = 2;
	string last_name = 3;
	Address address = 4;
	DOB date_of_birth = 5;
}

message FamilyMember {
	int32 age = 1;
	string sex = 2;
	bool spouse = 3;
}

message User {
	Identity identity = 1;
	repeated FamilyMember family_members = 2;
	string email_address = 3;
	string phone_number = 4;
	bool spanish = 5;
	string ulid = 6;
}`;

const BASE_64_REGEX = /^(?:[A-Za-z0-9+/]{4})*?(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/;
function is_proto_message( message ) {
	let b64_test = BASE_64_REGEX.test( message );
	if ( !b64_test ) { return false; }
	let proto_message = atob( message );
	console.log( proto_message );
	let proto_message_buffer = new Uint8Array( proto_message.length );
	for ( let i = 0; i < proto_message.length; i++ ) {
		proto_message_buffer[ i ] = proto_message.charCodeAt( i );
	}
	console.log( proto_message_buffer );
	var root = new protobuf.Root();
	protobuf.parse( USER_PROTO , root , { keepCase: true , alternateCommentMode: false , preferTrailingComment: false } );
	root.resolveAll();
	var UserMessage = root.lookupType( "user.User" );
	var decoded_message = UserMessage.decode( proto_message_buffer );
	console.log( "is_proto_message()" , decoded_message );
	return decoded_message
}

function set_cookie( name , value , days=3650 ) {
	let expires = "";
	let date = new Date();
	date.setTime( date.getTime() + ( days * 24 * 60 * 60 * 1000 ) );
	expires = "; expires=" + date.toUTCString();
	document.cookie = name + "=" + ( value || "" ) + expires + "; path=/; SameSite=Lax; Secure";
}

function get_cookie( name ) {
    let name_eq = name + "=";
    let ca = document.cookie.split( ";" );
    for( let i = 0; i < ca.length; i++ ) {
        let c = ca[ i ];
        while ( c.charAt( 0 ) === " " ) c = c.substring( 1 , c.length );
        if ( c.indexOf( name_eq ) === 0 ) return c.substring( name_eq.length , c.length );
    }
    return null;
}

function title_case( str ) {
	if ( !str ) { return ""; }
	return str.toLowerCase().split( " " ).map( x => {
		if ( !x ) { return ""; }
		return x[ 0 ].toUpperCase() + x.substr( 1 ).toLowerCase();
	}).join( " " ).trim();
}

function convert_milliseconds_to_time_string( milliseconds ) {
	let seconds = Math.floor( milliseconds / 1000 );
	let minutes = Math.floor( seconds / 60 );
	let hours = Math.floor( minutes / 60 );
	let days = Math.floor( hours / 24 );
	hours %= 24;
	minutes %= 60;
	seconds %= 60;

	let time_string = `${days} days , ${hours} hours , ${minutes} minutes , and ${seconds} seconds`;
	return time_string;
}

function set_nested_property( obj , keys , value ) {
	if ( keys.length === 1 ) {
		obj[ keys[ 0 ] ] = value;
	} else {
		const key = keys.shift();
		obj[ key ] = obj[ key ] || {};
		set_nested_property( obj[ key ] , keys , value );
	}
}

function add_qr_code( text , element_id ) {
	let x_element = document.getElementById( element_id );
	x_element.innerHTML = "";
	let user_qrcode = new QRCode( x_element , {
		text: text ,
		width: 256 ,
		height: 256 ,
		colorDark : "#000000" ,
		colorLight : "#ffffff" ,
		correctLevel : QRCode.CorrectLevel.H
	});
}

function set_url( new_url ) {
	// no page reload ?
	console.log( `Changing URL , FROM = ${window.location.href} || TO = ${new_url}` );
	window.history.pushState( null , null , new_url );

	// Update the query parameters
	// url.searchParams.set("q", "example");

	// Update the URL with a full page reload
	// window.location.href = url.toString();
}

function user_checkin_detect_uuid() {
	if ( !window.location?.href ) { return false; }
	let url_parts = window.location.href.split( "/checkin/" );
	if ( url_parts.length < 2 ) { return false; }
	if ( url_parts[ 1 ].length < 36 ) { return false; }
	let x_uuid = url_parts[ 1 ].substring( 0 , 36 );
	if ( is_uuid( x_uuid ) === false ) { return false; }
	return x_uuid;
}

function user_checkin_detect_state() {
	let url = window.location.href;
	if ( !url ) { return false; }
	let url_parts = window.location.href.split( "/" );
	if ( url_parts.length < 2 ) { return false; }
	if ( window.location.href.indexOf( "edit" ) > -1 ) {
		return "edit";
	}
	if ( window.location.href.indexOf( "new" ) > -1 ) {
		return "new";
	}
	return false;
}

function show_user_handoff_qrcode( x_uuid=false ) {
	if ( !x_uuid ) { x_uuid = window.USER.uuid; }
	let qr_code_link = `${window.location.protocol}//${window.location.host}/user/login/fresh/${x_uuid}`;
	add_qr_code( qr_code_link , "user-handoff-qr-code" );
	let user_handoff_modal = new bootstrap.Modal( "#user-handoff-modal" , {
		backdrop: "static" ,
		focus: true ,
		keyboard: true
	});
	user_handoff_modal.show();
}

function show_user_uuid_qrcode( x_uuid=false ) {
	if ( !x_uuid ) { x_uuid = window.USER.uuid; }
	add_qr_code( x_uuid , "user-handoff-qr-code" );
	let user_handoff_modal = new bootstrap.Modal( "#user-handoff-modal" , {
		backdrop: "static" ,
		focus: true ,
		keyboard: true
	});
	user_handoff_modal.show();
}