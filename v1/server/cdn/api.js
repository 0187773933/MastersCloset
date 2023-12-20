const ServerAPIKey = "";
const ServerBaseURL = "";
const LocalHostURL = "";

function api_check_in_uuid( uuid , balance_form_data ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let json_balance_form_data = JSON.stringify( balance_form_data );
			let check_in_url = `${ServerBaseURL}/admin/user/checkin/${uuid}`;
			let check_in_response = await fetch( check_in_url , {
				method: "POST" ,
				headers: { "key": ServerAPIKey } ,
				body: json_balance_form_data
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];

			// optionally enforce good result from synced-db ???
			// if ( result !== true ) { return; }
			// but then ....
			if ( LocalHostURL.length > 3 ) {
				console.log( "Sending Extra Print Request to Local Printer" );
				console.log( response_json[ "check_in" ][ "print_job" ] );
				let x_body = JSON.stringify( response_json[ "check_in" ][ "print_job" ] );
				console.log( x_body );
				// let check_in_response_print = await fetch( `${LocalHostURL}/admin/print` , {
				let check_in_response_print = await fetch( `${LocalHostURL}/admin/print2` , {
					method: "POST" ,
					headers: { "key": ServerAPIKey } ,
					body: x_body
				});
				let response_json_print = await check_in_response_print.json();
				// let result_print = response_json_print[ "result" ];
				console.log( response_json_print );
			}

			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_print( print_job ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let response_print = await fetch( `${LocalHostURL}/admin/print` , {
				method: "POST" ,
				headers: { "key": ServerAPIKey } ,
				body: JSON.stringify( print_job )
			});
			let response_json_print = await response_print.json();
			resolve( response_json_print );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_check_in_uuid_test( uuid ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_url = `${ServerBaseURL}/admin/user/checkin/test/${uuid}`;
			let check_in_response = await fetch( check_in_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			resolve( response_json );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_user_from_barcode( barcode ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_url = `${ServerBaseURL}/admin/user/get/barcode/${barcode}`;
			let check_in_response = await fetch( check_in_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let user = response_json[ "result" ];
			resolve( user );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_user_from_uuid( uuid ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_url = `${ServerBaseURL}/admin/user/get/${uuid}`;
			let check_in_response = await fetch( check_in_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let user = response_json[ "result" ];
			resolve( user );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_search_username( username ) {
	return new Promise( async function( resolve , reject ) {
		try {
			if ( !username ) { resolve( false ); return; }
			let search_url = `${ServerBaseURL}/admin/user/search/username/${username}`;
			let check_in_response = await fetch( search_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			if ( result === "not found" ) { result = false; }
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_fuzzy_search_username( username ) {
	return new Promise( async function( resolve , reject ) {
		try {
			if ( !username ) { resolve( false ); return; }
			let search_url = `${ServerBaseURL}/admin/user/search/username/fuzzy/${username}`;
			let check_in_response = await fetch( search_url , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_all_users() {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_response = await fetch( `${ServerBaseURL}/admin/user/get/all` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_edit_user( user_info ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let response = await fetch( `${ServerBaseURL}/admin/user/edit` , {
				method: "POST" ,
				body: JSON.stringify( user_info ) ,
				headers: { "key": ServerAPIKey }
			});
			let response_json = await response.json();
			resolve( response_json );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_all_emails() {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_response = await fetch( `${ServerBaseURL}/admin/user/get/all/emails` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_all_barcodes() {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_response = await fetch( `${ServerBaseURL}/admin/user/get/all/barcodes` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_all_checkins() {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_response = await fetch( `${ServerBaseURL}/admin/user/get/all/checkins` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_all_checkins_for_date( date_key ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_response = await fetch( `${ServerBaseURL}/admin/checkins/get/${date_key}` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_all_phone_numbers() {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_response = await fetch( `${ServerBaseURL}/admin/user/get/all/phone-numbers` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_new_user( user_info ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let response = await fetch( `${ServerBaseURL}/admin/user/new` , {
				method: "POST" ,
				body: JSON.stringify( user_info ) ,
				headers: { "key": ServerAPIKey }
			});
			let response_json = await response.json();
			resolve( response_json );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_delete_user( uuid ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let response = await fetch( `${ServerBaseURL}/admin/user/delete/${uuid}` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await response.json();
			resolve( response_json );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_delete_checkin( uuid , ulid ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let response = await fetch( `${ServerBaseURL}/admin/checkins/delete/${uuid}/${ulid}` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await response.json();
			resolve( response_json );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_submit_form( url , form_data ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let response = await fetch( `${ServerBaseURL}/${url}` , {
				method: "POST" ,
				body: form_data ,
				headers: { "key": ServerAPIKey }
			});
			let response_json = await response.json();
			if ( !response_json[ "result" ] ) { resolve( false ); return; }
			resolve( response_json[ "result" ] );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_all_log_files() {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_response = await fetch( `${ServerBaseURL}/admin/logs/get/log-file-names` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_log_file( file_path ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_response = await fetch( `${ServerBaseURL}/admin/logs/get/${file_path}` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}

function api_get_check_in( uuid , ulid ) {
	return new Promise( async function( resolve , reject ) {
		try {
			let check_in_response = await fetch( `${ServerBaseURL}/admin/checkins/get/${uuid}/${ulid}` , {
				method: "GET" ,
				headers: { "Content-Type": "application/json" , "key": ServerAPIKey }
			});
			let response_json = await check_in_response.json();
			let result = response_json[ "result" ];
			resolve( result );
			return;
		}
		catch( error ) { console.log( error ); resolve( false ); return; }
	});
}