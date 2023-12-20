const uuid_v4_regex = /^[0-9A-F]{8}-[0-9A-F]{4}-[4][0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12}$/i
function is_uuid( str ) { return uuid_v4_regex.test( str ); }
const barcode_regex = /^\d+$/;
function is_barcode( str ) { return barcode_regex.test( str ); }
function sleep( ms ) { return new Promise( resolve => setTimeout( resolve , ms ) ); }

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