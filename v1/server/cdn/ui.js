function get_ui_user_qr_code_display() {
	return `
	<div class="row">
		<div class="col-md-6">
			<div id="user-handoff-modal" class="modal fade" data-bs-backdrop="static" data-bs-keyboard="false" tabindex="-1" aria-labelledby="staticBackdropLabel" aria-hidden="true">
				<div class="modal-dialog modal-dialog-centered modal-dialog-scrollable" >
					<div class="modal-dialog">
						<div class="modal-content bg-success-subtle">
							<div class="modal-header">
								<h5 class="modal-title">Masters Closet Login</h5>
								<button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
							</div>
							<div class="modal-body">
								<!-- <p>Please take a picture of this QR Code to Login Next Time</p> -->
								<center>
									<div id="user-handoff-qr-code"></div>
								</center>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>`;
}

function get_ui_alert_check_in_allowed() {
	return `
	<div class="row">
		<div class="col-md-4"></div>
		<div class="col-md-4">
			<div class="alert alert-success" id="checked-in-alert-true">
				<center>Allowed to Check In !!!</center>
			</div>
		</div>
		<div class="col-md-4"></div>
	</div>`;
}

function get_ui_alert_check_in_failed() {
	return `
	<div class="row">
		<div class="col-md-3"></div>
		<div class="col-md-6">
			<div class="alert alert-danger" id="checked-in-alert-false">
				<center>
					Checked In Too Recently !!!<br><br>
					<a id="block-button" class="btn btn-warning" target="_blank" href="/none">Block</a>
				</center>
			</div>
		</div>
		<div class="col-md-3"></div>
	</div>`;
}

function get_ui_active_user_info() {
	return `
	<div class="row">
		<center><h2 id="active-username"></h2></center>
		<center><h4 id="active-user-time-remaining"></h4></center>
	</row>
	`;
}

// TODO : Make shopping_for settable in config.json
function get_ui_shopping_for_selector() {
	return `
	<div class="row">
		<div class="col-md-3"></div>
		<div class="col-md-6">
			<div class="input-group">
				<div class="input-group-text">Shopping For</div>
				<select id="shopping_for" class="form-select" aria-label="Shopping For" name="shopping_for">
					<option value="1">1</option>
					<option value="2">2</option>
					<option value="3">3</option>
					<option value="4">4</option>
					<option value="5">5</option>
				</select>
			</div>
		</div>
		<div class="col-md-3"></div>
	</div>
	`;
}

function get_ui_shopping_for_selector_advanced() {
	return `
	<div class="row">
		<div class="col-lg-2"></div>
		<div class="col-lg-2">
			<h3>Shopping For :</h3>
		</div>
		<div class="col-sm-12 col-md-4 col-lg-6">
			<div id="user_family_members"></div>
		</div>
		<div class="col-lg-2"></div>
	</div>
	<br>
	`;
}

function get_ui_user_search_table() {
	return `
	<div class="row">
		<div class="col-md-1"></div>
		<div class="col-md-10">
			<div class="table-responsive-sm">
				<table id="user-search-table" class="table table-hover table-striped-columns">
					<thead>
						<tr>
							<th scope="col">#</th>
							<th scope="col">Username</th>
							<th scope="col">UUID</th>
							<th scope="col">Select</th>
						</tr>
					</thead>
					<tbody id="user-search-table-body"></tbody>
				</table>
			</div>
		</div>
		<div class="col-md-1"></div>
	</div>`;
}
function populate_user_search_table( users ) {
	// console.log( "populate_user_search_table()" );
	// console.log( users );
	$( "#user-search-table" ).show();
	let table_body_element = document.getElementById( "user-search-table-body" );
	table_body_element.innerHTML = "";
	for ( let i = 0; i < users.length; ++i ) {
		let _tr = document.createElement( "tr" );

		let user_number = document.createElement( "th" );
		user_number.setAttribute( "scope" , "row" );
		user_number.textContent = `${(i + 1)}`;
		_tr.appendChild( user_number );

		let username = document.createElement( "td" );
		username.textContent = users[ i ][ "username" ];
		_tr.appendChild( username );

		let uuid_holder = document.createElement( "td" );
		let uuid_text = document.createElement( "span" );
		uuid_text.textContent = users[ i ][ "uuid" ];
		uuid_text.innerHTML += "&nbsp;&nbsp;"
		uuid_holder.appendChild( uuid_text );
		_tr.appendChild( uuid_holder );

		let select_button_holder = document.createElement( "td" );
		let select_button = document.createElement( "button" );
		select_button.textContent = "Select"
		select_button.className = "btn btn-success btn-sm";
		select_button.onclick = function() {
			$( "#user-search-input" ).val( users[ i ][ "uuid" ] );
			// $( "#user-search-table" ).hide();
			// check_in_uuid_input();
			// search_input();
			window.USER = users[ i ];
			// _on_check_in_input_change( users[ i ][ "uuid" ] );
			// $( "#main-row" ).trigger( "render_active_user" , users[ i ] );
			window.UI.render_active_user();
		};
		select_button_holder.appendChild( select_button );
		_tr.appendChild( select_button_holder );

		table_body_element.appendChild( _tr );
	}
}

function get_ui_user_balance_table() {
	return `
	<div class="row">
		<div class="col-md-1"></div>
		<div class="col-md-10">
			<div class="table-responsive-sm">
				<table id="user-balance-table" class="table table-hover table-striped-columns">
					<thead>
						<tr>
							<th scope="col">Item</th>
							<th scope="col">Available</th>
							<!-- <th scope="col">Limit</th> -->
							<th scope="col">Total Used</th>
						</tr>
					</thead>
					<tbody id="user-balance-table-body"></tbody>
				</table>
			</div>
		</div>
		<div class="col-md-1"></div>
	</div>`;

}
function _add_balance_row( table_body_element , name , available , limit , used ) {
	let _tr = document.createElement( "tr" );
	let item = document.createElement( "th" );
	item.textContent = name;
	_tr.appendChild( item );
	let _available = document.createElement( "td" );
	let available_input = document.createElement( "input" );
	available_input.setAttribute( "type" , "text" );
	available_input.className = "form-control";
	available_input.value = available;
	available_input.setAttribute( "id" , `balance_${name.toLowerCase()}_available` );
	_available.appendChild( available_input );
	_tr.appendChild( _available );
	// let _limit = document.createElement( "td" );
	// let limit_input = document.createElement( "input" );
	// limit_input.setAttribute( "type" , "text" );
	// limit_input.className = "form-control";
	// limit_input.value = limit;
	// limit_input.setAttribute( "id" , `balance_${name.toLowerCase()}_limit` );
	// limit_input.setAttribute( "readonly" , "" );
	// _limit.appendChild( limit_input );
	// _tr.appendChild( _limit );
	let _used = document.createElement( "td" );
	let used_input = document.createElement( "input" );
	used_input.setAttribute( "type" , "text" );
	used_input.className = "form-control";
	used_input.value = used;
	used_input.setAttribute( "id" , `balance_${name.toLowerCase()}_used` );
	used_input.setAttribute( "readonly" , "" );
	_used.appendChild( used_input );
	_tr.appendChild( _used );
	table_body_element.appendChild( _tr );
}

// could just switch to multiple inputs ?
// https://getbootstrap.com/docs/5.3/forms/input-group/#multiple-inputs
function populate_user_balance_table( shopping_for , balance , balance_config ) {

	console.log( "populate_user_balance_table()" );
	console.log( "shopping for === " , shopping_for );
	console.log( "balance === " , balance );
	console.log( "balance_config === " , balance_config );

	let tops_available = ( shopping_for * balance_config.general.tops );
	let bottoms_available = ( shopping_for * balance_config.general.bottoms );
	let dresses_available = ( shopping_for * balance_config.general.dresses );
	let shoes_available = ( shopping_for * balance_config.shoes );
	let seasonal_available = ( shopping_for * balance_config.seasonals );
	let accessories_available = ( shopping_for * balance_config.accessories );

	let table_body_element = document.getElementById( "user-balance-table-body" );
	table_body_element.innerHTML = "";

	_add_balance_row( table_body_element , "Tops" ,
		tops_available ,
		balance[ "general" ][ "tops" ][ "limit" ] ,
		balance[ "general" ][ "tops" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Bottoms" ,
		bottoms_available ,
		balance[ "general" ][ "bottoms" ][ "limit" ] ,
		balance[ "general" ][ "bottoms" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Dresses" ,
		dresses_available ,
		balance[ "general" ][ "dresses" ][ "limit" ] ,
		balance[ "general" ][ "dresses" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Shoes" ,
		shoes_available ,
		balance[ "shoes" ][ "limit" ] ,
		balance[ "shoes" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Seasonals" ,
		seasonal_available ,
		balance[ "seasonals" ][ "limit" ] ,
		balance[ "seasonals" ][ "used" ] ,
	);

	_add_balance_row( table_body_element , "Accessories" ,
		accessories_available ,
		balance[ "accessories" ][ "limit" ] ,
		balance[ "accessories" ][ "used" ] ,
	);

}

function _get_user_form() {
	return `
				<!-- Main Required Stuff -->
				<div class="row g-2 mb-3">
					<div class="col-lg-2"></div>
					<div class="col-sm-12 col-md-4 col-lg-3">
						<div class="form-floating">
							<input id="user_first_name" type="text" class="form-control input-name" name="user_first_name">
							<label for="user_first_name">First Name</label>
						</div>
					</div>
					<div class="col-sm-12 col-md-4 col-lg-3">
						<div class="form-floating">
							<input id="user_middle_name" type="text" class="form-control input-name" name="user_middle_name">
							<label for="user_middle_name">Middle Name</label>
						</div>
					</div>
					<div class="col-sm-12 col-md-4 col-lg-3">
						<div class="form-floating">
							<input id="user_last_name" type="text" class="form-control input-name" name="user_last_name">
							<label for="user_last_name">Last Name</label>
						</div>
					</div>
					<div class="col-lg-2"></div>
				</div>

				<!-- Contact Info -->
				<div class="row g-2 mb-3">
					<div class="col-lg-3"></div>
					<div class="col-sm-12 col-md-4 col-lg-4">
						<div class="form-floating">
							<input id="user_email" type="email" class="form-control" name="user_email">
							<label for="user_email">Email Address</label>
						</div>
					</div>
					<div class="col-sm-12 col-md-4 col-lg-3">
						<div class="form-floating">
							<input id="user_phone_number" type="tel" class="form-control" name="user_phone_number">
							<label for="user_phone_number">Phone Number</label>
						</div>
					</div>
					<div class="col-lg-3"></div>
				</div>

				<!-- Barcodes -->
				<div class="row g-2 mb-3">
					<div class="col-lg-4 col-md-4"></div>
					<div class="col-lg-4 col-md-4 col-sm-12">
						<center><button id="add-barcode-button" class="btn btn-qrcode" onclick="on_add_barcode(event);">Add Barcode</button></center>
					</div>
					<div class="col-lg-4 col-md-4"></div>
				</div>
				<div class="row g-2 mb-3">
					<div class="col-lg-3 col-md-0 col-sm-0 col-0"></div>
					<div class="col-lg-6 col-md-12 col-sm-12 col-12">
						<div id="user_barcodes"></div>
					</div>
					<div class="col-lg-3 col-md-0 col-sm-0 col-0"></div>
				</div>

				<!-- Family Members -->
				<div class="row g-2 mb-3">
					<div class="col-md-4"></div>
					<div class="col-md-4">
						<center><button id="add-family-member-button" class="btn btn-primary" onclick="on_add_family_member(event);">Add Family Member</button></center>
					</div>
					<div class="col-md-4"></div>
				</div>
				<div id="user_family_members"></div>

				<!-- Address - Part 1-->
				<div class="row g-2 mb-3">
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_street_number" type="text" class="form-control" name="user_street_number">
							<label for="user_street_number">Street Number</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_street_name" type="text" class="form-control" name="user_street_name">
							<label for="user_street_name">Street Name</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_address_two" type="text" class="form-control" name="user_street_name">
							<label for="user_address_two">Address 2</label>
						</div>
					</div>
				</div>

				<!-- Address - Part 2-->
				<div class="row g-2 mb-3">
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_city" type="text" class="form-control" name="user_city">
							<label for="user_city">City</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_state" type="text" class="form-control" name="user_state">
							<label for="user_state">State</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_zip_code" type="text" class="form-control" name="user_zip_code">
							<label for="user_zip_code">Zip Code</label>
						</div>
					</div>
				</div>

				<!-- Extras -->
				<div class="row g-2 mb-3">

					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_birth_day" type="number" min="1" max="31" class="form-control" name="user_birth_day_name">
							<label for="user_birth_day_name">Birth Day</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<select id="user_birth_month" class="form-select" aria-label="User Birth Month" name="user_birth_month">
								<option value="JAN">JAN = 1</option>
								<option value="FEB">FEB = 2</option>
								<option value="MAR">MAR = 3</option>
								<option value="APR">APR = 4</option>
								<option value="MAY">MAY = 5</option>
								<option value="JUN">JUN = 6</option>
								<option value="JUL">JUL = 7</option>
								<option value="AUG">AUG = 8</option>
								<option value="SEP">SEP = 9</option>
								<option value="OCT">OCT = 10</option>
								<option value="NOV">NOV = 11</option>
								<option value="DEC">DEC = 12</option>
							</select>
							<label for="user_birth_month">Birth Month</label>
						</div>
					</div>
					<div class="col-md-4">
						<div class="form-floating">
							<input id="user_birth_year" type="number" min="1900" max="2100" class="form-control" name="user_birth_year_name">
							<label for="user_birth_year_name">Birth Year</label>
						</div>
					</div>
				</div>

				<div class="row g-2 mb-3">
					<div class="col-md-5 col-lg-5"></div>
					<div class="col-md-2 col-lg-2">
						<div class="form-check">
							<input class="form-check-input" type="checkbox" value="" id="user_spanish">
							<label class="form-check-label" for="user_spanish">Español</label>
						</div>
					</div>
					<div class="col-md-5 col-lg-5"></div>
				</div>
	`;
}

// function get_ui_user_new_form() {
// 	return `
// 	<div class="row">
// 		<center>
// 			<form id="user-new-form" action="/admin/user/new" method="post">
// 				${_get_user_form()}
// 			</form>
// 		</center>
// 	</div>`;
// }

function get_ui_user_new_form() {
	return `
	<div class="row">
		<form id="user-new-form" action="/admin/user/new" method="post">
			${_get_user_form()}
		</form>
	</div>`;
}

// function get_ui_user_edit_form() {
// 	return `
// 	<div class="row">
// 		<center>
// 			<form id="user-new-form" action="/admin/user/edit" method="post">
// 				${_get_user_form()}
// 			</form>
// 		</center>
// 	</div>`;
// }

function get_ui_user_edit_form() {
	return `
	<div class="row">
		<form id="user-new-form" action="/admin/user/edit" method="post">
			${_get_user_form()}
		</form>
	</div>`;
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
function show_user_exists_modal( uuid ) {
	let qr_code_link = `${window.location.protocol}//${window.location.host}/user/login/fresh/${uuid}`;
	add_qr_code( qr_code_link , "user-exists-qr-code" );
	let user_exists_modal = new bootstrap.Modal( "#user-exists-error-modal" , {
		backdrop: "static" ,
		focus: true ,
		keyboard: true
	});
	user_exists_modal.show();
}
function show_user_handoff_modal( uuid ) {
	let qr_code_link = `${window.location.protocol}//${window.location.host}/user/login/fresh/${uuid}`;
	add_qr_code( qr_code_link , "user-handoff-qr-code" );
	let user_handoff_modal = new bootstrap.Modal( "#user-handoff-modal" , {
		backdrop: "static" ,
		focus: true ,
		keyboard: true
	});
	user_handoff_modal.show();
}


function make( element_type , options ) {
	let element = document.createElement( element_type );
	if ( options.class ) { element.className = options.class; }
	if ( options.text ) { element.textContent = options.text; }
	if ( options.style ) { element.style = options.style; }
	if ( options.id ) { element.id = options.id; }
	if ( options.attributes ) {
		let attributes = Object.keys( options.attributes );
		for ( let i = 0; i < attributes.length; ++i ) {
			element.setAttribute( attributes[ i ] , options.attributes[ attributes[ i ] ] );
		}
	}
	if ( options.listeners ) {
		let listeners = Object.keys( options.listeners );
		for ( let i = 0; i < listeners.length; ++i ) {
			element.addEventListener( listeners[ i ] , options.listeners[ listeners[ i ] ] );
		}
	}
	return element;
}

function on_add_family_member( event ) {
	if ( event ) { event.preventDefault(); }
	console.log( "on_add_family_member()" );
	let family_member_ulid = ULID.ulid();
	let family_member_id = `user_family_member_${family_member_ulid}`;
	window.FAMILY_MEMBERS[ family_member_ulid ] = { "age": -1 , "spouse": false , "sex": "" };
	let current_family_members = document.querySelectorAll( ".user-family-member" );
	if ( current_family_members.length >= 5 ) { return; }
	let holder = document.getElementById( "user_family_members" );

	// 1.) Label Row
	let label_row = make( "div" , {
		id: `user_family_member_row_label_${family_member_ulid}` ,
		class: "row g-2 mb-3 justify-content-center" ,
	});
	let label_col_left = make( "div" , {
		class: "col-lg-2 col-md-2 col-sm-2 col-5"
	});
	let label_col_center = make( "div" , {
		class: "col-lg-2 col-md-2 col-sm-2 col-2"
	});
	let label_col_right = make( "div" , {
		class: "col-lg-2 col-md-2 col-sm-2 col-5"
	});
	let label = make( "span" , {
		class: "badge rounded-pill text-bg-primary" ,
		text: `Family Member - ${(current_family_members.length + 1)}` ,
		attributes: {
			id: `user_family_member_label_${family_member_ulid}`
		}
	});
	let x_center = make( "center" , {} );
	x_center.appendChild( label );
	label_col_center.appendChild( x_center );
	label_row.appendChild( label_col_left );
	label_row.appendChild( label_col_center );
	label_row.appendChild( label_col_right );
	holder.appendChild( label_row );

	// 2.) Everything Else Row
	let new_row = make( "div" , {
		id: `user_family_member_row_${family_member_ulid}` ,
		class: "row g-2 mb-3 justify-content-center" ,
	});
	let new_row_line_break = make( "br" , {} );
	new_row.appendChild( new_row_line_break );

	// Age
	let age_col = make( "div" , { class: "col-lg-1 col-md-2 col-sm-3 col-3" });
	let age_form = make( "div" , { class: "form-floating" });
	let age_input = make( "input" , {
		class: "form-control user-family-member" ,
		attributes: {
			type: "number" ,
			id: `user_family_member_${family_member_ulid}_age`
		} ,
		listeners: {
			keydown: ( event ) => { if ( event.keyCode === 13 ) { event.preventDefault(); } } ,
			keyup: ( event ) => {
				window.FAMILY_MEMBERS[ family_member_ulid ].age = event.target.value;
				let male_text = document.getElementById( `user_family_member_${family_member_ulid}_gender_label_male` );
				let female_text = document.getElementById( `user_family_member_${family_member_ulid}_gender_label_female` );
				let spouse_button = document.getElementById( `user_family_member_${family_member_ulid}_spouse` );
				if ( event.target.value < 18 ) {
					male_text.textContent = "Boy";
					female_text.textContent = "Girl";
					spouse_button.parentNode.parentNode.style.display = "none";
				} else {
					male_text.textContent = "Male";
					female_text.textContent = "Female";
					spouse_button.parentNode.parentNode.style.display = "block";
				}
			}
		}
	});
	let age_label = make ( "label" , {
		text: "Age" ,
		attributes: { for: `user_family_member_${family_member_ulid}_age` }
	});
	age_form.appendChild( age_input );
	age_form.appendChild( age_label );
	age_col.appendChild( age_form );
	new_row.appendChild( age_col );

	// Gender
	let gender_col = make( "div" , {
		class: "col-lg-1 col-md-1 col-sm-3 col-3" ,
		style: "padding-left: 1em; padding-right: 0em;"
	});
	let male_gender_form = make( "div" , { class: "form-check" });
	let male_gender_input = make( "input" , {
		class: "form-check-input" ,
		attributes: {
			type: "radio" ,
			id: `user_family_member_${family_member_ulid}_gender_male` ,
			name: `user_family_member_${family_member_ulid}_gender`
		} ,
		listeners: {
			change: ( event ) => {
				if ( event.target.checked ) {
					window.FAMILY_MEMBERS[ family_member_ulid ].sex = "male";
					console.log( "changed sex to male" , window.FAMILY_MEMBERS[ family_member_ulid ] );
				}
			}
		}
	});
	let male_gender_label = make( "label" , {
		text: "Male" ,
		class: "form-check-label" ,
		attributes: {
			for: `user_family_member_${family_member_ulid}_gender_male` ,
			id: `user_family_member_${family_member_ulid}_gender_label_male` ,
		}
	});
	male_gender_form.appendChild( male_gender_input );
	male_gender_form.appendChild( male_gender_label );
	gender_col.appendChild( male_gender_form );

	let female_gender_form = make( "div" , { class: "form-check" });
	let female_gender_input = make( "input" , {
		class: "form-check-input" ,
		attributes: {
			type: "radio" ,
			id: `user_family_member_${family_member_ulid}_gender_female` ,
			name: `user_family_member_${family_member_ulid}_gender`
		} ,
		listeners: {
			change: ( event ) => {
				if ( event.target.checked ) {
					window.FAMILY_MEMBERS[ family_member_ulid ].sex = "female";
					console.log( "changed sex to female" , window.FAMILY_MEMBERS[ family_member_ulid ] );
				}
			}
		}
	});
	let female_gender_label = make( "label" , {
		text: "Female" ,
		class: "form-check-label" ,
		attributes: {
			for: `user_family_member_${family_member_ulid}_gender_female` ,
			id: `user_family_member_${family_member_ulid}_gender_label_female` ,
		}
	});
	female_gender_form.appendChild( female_gender_input );
	female_gender_form.appendChild( female_gender_label );
	gender_col.appendChild( female_gender_form );

	new_row.appendChild( gender_col );

	// Spouse
	let col_5 = document.createElement( "div" );
	col_5.className = "col-lg-1 col-md-3 col-sm-3 col-3";
	let spouse_form = document.createElement( "div" );
	spouse_form.className = "form-check-reverse form-switch";
	let spouse_input = document.createElement( "input" );
	spouse_input.className = "form-check-input";
	spouse_input.setAttribute( "type" , "checkbox" );
	spouse_input.setAttribute( "role" , "switch" );
	spouse_input.setAttribute( "id" , `user_family_member_${family_member_ulid}_spouse` );
	spouse_input.setAttribute( "name" , `user_family_member_${family_member_ulid}_spouse` );
	spouse_input.addEventListener( "change" , function( event ) {
		if ( event.target.checked ) {
			window.FAMILY_MEMBERS[ family_member_ulid ].spouse = true;
			console.log( "selected spouse" , window.FAMILY_MEMBERS[ family_member_ulid ] );
		} else {
			window.FAMILY_MEMBERS[ family_member_ulid ].spouse = false;
			console.log( "de-selected spouse" , window.FAMILY_MEMBERS[ family_member_ulid ] );
		}
	});
	let spouse_label = document.createElement( "label" );
	spouse_label.className = "form-check-label";
	spouse_label.setAttribute( "for" , `user_family_member_${family_member_ulid}_spouse` );
	spouse_label.checked = true;
	spouse_label.textContent = "Spouse";
	spouse_form.appendChild( spouse_input );
	spouse_form.appendChild( spouse_label );
	col_5.appendChild( spouse_form );
	let col_5_span = document.createElement( "span" );
	col_5_span.textContent = "  ";
	col_5.appendChild( col_5_span );
	new_row.appendChild( col_5 );

	let col_6 = document.createElement( "div" );
	col_6.className = "col-lg-1 col-md-2 col-sm-2 col-2";
	let delete_button = document.createElement( "a" );
	delete_button.className = "btn btn-danger p-1";
	let delete_button_icon = document.createElement( "i" );
	delete_button_icon.className = "bi bi-trash3-fill";
	delete_button.appendChild( delete_button_icon );
	delete_button.onclick = function( event ) {
		if ( event ) { event.preventDefault(); }

		console.log( `delete_button.onclick( ${family_member_ulid} )` );
		// let x_id = event?.target?.parentNode?.parentNode?.id;
		// if ( x_id === undefined ) { x_id = event?.target?.parentNode?.parentNode?.parentNode?.id; }
		// if ( x_id === "" ) { x_id = event?.target?.parentNode?.parentNode?.parentNode?.id; }
		// if ( x_id === "" ) { console.log( event?.target?.parentNode?.parentNode ); }
		// let x_id_parts = x_id.split( "_" );
		// x_id = x_id_parts[ x_id_parts.length - 1 ];
		// console.log( "x_id === " , x_id );
		let result = confirm( `Are You Absolutely Sure You Want to Delete This Family Member ???` );
		if ( result === true ) {

			console.log( `Deleting ${family_member_ulid}` );

			delete window.FAMILY_MEMBERS[ family_member_ulid ];

			$( `#user_family_member_row_label_${family_member_ulid}` ).remove();
			$( `#user_family_member_row_${family_member_ulid}` ).remove();

			// Update Numbers
			let labels = document.querySelectorAll( '[id^="user_family_member_label"]' );
			for ( let i = 0; i < labels.length; ++i ) {
				labels[ i ].innerText = `Family Member - ${(i+1)}`;
			}
			return;
		}
	};
	col_6.appendChild( delete_button );
	new_row.appendChild( col_6 );

	if ( current_family_members.length === 0 ) {
		new_row.setAttribute( "style" , "padding-bottom: 2px;" );
	}

	let n_center = make( "center" , {} );
	n_center.appendChild( new_row );
	holder.appendChild( n_center );
	document.getElementById( `user_family_member_${family_member_ulid}_age` ).focus();
	return family_member_ulid;
}

// user_family_members
function on_add_family_member_display( event ) {
	if ( event ) { event.preventDefault(); }
	console.log( "on_add_family_member_display()" );
	let current_family_members = document.querySelectorAll( ".user-family-member" );
	if ( current_family_members.length >= 6 ) { return; }
	let family_member_ulid = ULID.ulid();
	let family_member_id = `user_family_member_${family_member_ulid}`;
	window.FAMILY_MEMBERS[ family_member_ulid ] = { "age": -1 , "spouse": false , "sex": "" };
	let holder = document.getElementById( "user_family_members" ); // this is a bootstrap column
	let new_row = document.createElement( "div" );
	new_row.setAttribute( "id" , `user_family_member_row_${family_member_ulid}` );
	new_row.className = "row no-gutters";
	let line_break = document.createElement( "br" );
	new_row.appendChild( line_break );

	let container = document.createElement( "div" );
	container.className = "d-flex justify-content-start";

	// Un-Named Family Member Name/ID
	let col_2 = document.createElement( "div" );
	col_2.className = "p-2";
	let name = document.createElement( "button" );
	name.setAttribute( "type" , "button" );
	name.className = "btn btn-primary user-family-member";
	name.textContent = `Family Member - ${(current_family_members.length)}`; // take off +1 because of self add
	name.setAttribute( "id" , `user_family_member_label_${family_member_ulid}` );
	name.addEventListener( "click" , function( event ) {
		if ( event.target.classList.contains( "btn-primary" ) ) {
			event.target.classList.remove( "btn-primary" );
			event.target.classList.add( "btn-light" );
		} else {
			event.target.classList.remove( "btn-light" );
			event.target.classList.add( "btn-primary" );
		}
		let shopping_for = document.querySelectorAll( ".user-family-member.btn-primary" ).length;
		populate_user_balance_table( shopping_for , window.USER.balance , window.BALANCE_CONFIG );
	});
	col_2.appendChild( name );
	container.appendChild( col_2 );

	// Age
	let col_3 = document.createElement( "div" );
	col_3.className = "p-2";
	let age_holder = document.createElement( "h4" );
	let age_badge = document.createElement( "span" );
	age_badge.setAttribute( "id" , `user_family_member_${family_member_ulid}_age` );
	age_badge.className = "badge bg-dark";
	age_badge.textContent = "";
	age_holder.appendChild( age_badge );
	col_3.appendChild( age_holder );
	container.appendChild( col_3 );

	// Gender
	let col_4 = document.createElement( "div" );
	col_4.className = "p-2";
	let gender_holder = document.createElement( "h4" );
	let gender_badge = document.createElement( "span" );
	gender_badge.setAttribute( "id" , `user_family_member_${family_member_ulid}_gender_text` );
	gender_badge.className = "badge bg-secondary";
	gender_badge.textContent = "";
	gender_holder.appendChild( gender_badge );
	col_4.appendChild( gender_holder );
	container.appendChild( col_4 );

	// Spouse
	let col_5 = document.createElement( "div" );
	col_5.className = "p-2";
	let spouse_holder = document.createElement( "h4" );
	let spouse_badge = document.createElement( "span" );
	spouse_badge.setAttribute( "id" , `user_family_member_${family_member_ulid}_spouse` );
	spouse_badge.className = "badge bg-primary";
	spouse_badge.setAttribute( "style" , "background-color: #A76385 !important; display: none;" );
	spouse_badge.textContent = "Spouse";
	spouse_holder.appendChild( spouse_badge );
	col_5.appendChild( spouse_holder );
	container.appendChild( col_5 );

	new_row.appendChild( container );
	new_row.setAttribute( "style" , "padding-bottom: 5px;" );

	holder.appendChild( new_row );
	return family_member_ulid;
}

function on_add_barcode( event ) {
	if ( event ) { event.preventDefault(); }
	console.log( "on_add_barcode()" );
	let barcode_ulid = ULID.ulid();
	let barcode_id = `user_barcode_${barcode_ulid}`;
	window.BARCODES[ barcode_id ] = "";
	let current_barcodes = document.querySelectorAll( ".user-barcode" );
	let holder = document.getElementById( "user_barcodes" );

	let new_row = document.createElement( "div" );
	new_row.setAttribute( "id" , `user_barcode_row_${barcode_ulid}` );
	new_row.className = "row g-2";

	let col_1 = document.createElement( "div" );
	col_1.className = "col-md-3";
	new_row.appendChild( col_1 );

	let col_2 = document.createElement( "div" );
	col_2.className = "col-md-6";
	let input_group = document.createElement( "div" );
	input_group.className = "input-group";
	let label = document.createElement( "span" );
	label.className = "input-group-text";
	label.setAttribute( "id" , `user_barcode_label_${barcode_ulid}` );
	label.textContent = `Barcode - ${(current_barcodes.length + 1)}`;
	let barcode_input = document.createElement( "input" );
	barcode_input.className = "form-control user-barcode";
	barcode_input.setAttribute( "placeholder" , "Barcode Number" );
	barcode_input.setAttribute( "type" , "text" );
	barcode_input.setAttribute( "name" , barcode_id );
	barcode_input.setAttribute( "id" , barcode_id );
	barcode_input.addEventListener( "keydown" , ( event ) => {
		if ( event.keyCode === 13 ) {
			event.preventDefault();
			return;
		}
	});
	barcode_input.addEventListener( "keyup" , ( event ) => {
		// window.USER.barcodes[ current_barcodes.length ] = event.target.value;
		window.BARCODES[ barcode_ulid ] = event.target.value;
	});

	input_group.appendChild( label );
	input_group.appendChild( barcode_input );

	let barcode_delete_button = document.createElement( "a" );
	barcode_delete_button.className = "btn btn-danger p-1 d-flex justify-content-center align-items-center";
	let barcode_delete_button_icon = document.createElement( "i" );
	barcode_delete_button_icon.className = "bi bi-trash3-fill";
	barcode_delete_button.appendChild( barcode_delete_button_icon );
	barcode_delete_button.onclick = async function( event ) {
		if ( event ) { event.preventDefault(); }
		let barcode_id = event?.target?.parentNode?.parentNode?.childNodes[ 1 ]?.id;
		if ( barcode_id === undefined ) { bardcode_id = event?.target?.parentNode?.childNodes[ 1 ]?.id; }
		if ( barcode_id === undefined ) { console.log( event.target ); }
		let barcode_ulid = barcode_id.split( "user_barcode_" )[ 1 ];
		let result = confirm( `Are You Absolutely Sure You Want to Delete This Barcode ???` );
		if ( result === true ) {
			delete window.BARCODES[ barcode_ulid ];
			let row_id = `#user_barcode_row_${barcode_ulid}`;
			$( row_id ).remove();
			let labels = document.querySelectorAll( '[id^="user_barcode_label_"]' );
			for ( let i = 0; i < labels.length; ++i ) {
				console.log( labels[ i ].innerText , `Barcode - ${(i+1)}` );
				labels[ i ].innerText = `Barcode - ${(i+1)}`;
			}
			return;
		}
	};
	input_group.appendChild( barcode_delete_button );
	col_2.appendChild( input_group );
	// col_2.appendChild( barcode_delete_button );

	new_row.appendChild( col_2 );

	let col_3 = document.createElement( "div" );
	col_3.className = "col-md-3";
	new_row.appendChild( col_3 );

	holder.appendChild( new_row );
	document.getElementById( barcode_id ).focus();
	return barcode_ulid;
}
function populate_user_edit_form( user_info ) {
	console.log( "populate_user_edit_form()" );
	console.log( user_info );
	window.BARCODES = {};
	window.FAMILY_MEMBERS = {};
	// console.log( JSON.stringify( user_info , null , 4 ) );
	let first_name_element = document.getElementById( "user_first_name" );
	first_name_element.value = user_info[ "identity" ][ "first_name" ];
	let middle_name_element = document.getElementById( "user_middle_name" );
	middle_name_element.value = user_info[ "identity" ][ "middle_name" ];
	let last_name_element = document.getElementById( "user_last_name" );
	last_name_element.value = user_info[ "identity" ][ "last_name" ];
	let email_element = document.getElementById( "user_email" );
	email_element.value = user_info[ "email_address" ];
	let phone_number_element = document.getElementById( "user_phone_number" );
	phone_number_element.value = user_info[ "phone_number" ];
	let street_number_element = document.getElementById( "user_street_number" );
	street_number_element.value = user_info[ "identity" ][ "address" ][ "street_number" ];
	let street_name_element = document.getElementById( "user_street_name" );
	street_name_element.value = user_info[ "identity" ][ "address" ][ "street_name" ];
	let address_two_element = document.getElementById( "user_address_two" );
	address_two_element.value = user_info[ "identity" ][ "address" ][ "address_two" ];
	let city_element = document.getElementById( "user_city" );
	city_element.value = user_info[ "identity" ][ "address" ][ "city" ];
	let state_element = document.getElementById( "user_state" );
	state_element.value = user_info[ "identity" ][ "address" ][ "state" ];
	let zip_code_element = document.getElementById( "user_zip_code" );
	zip_code_element.value = user_info[ "identity" ][ "address" ][ "zipcode" ];

	if ( user_info[ "identity" ][ "date_of_birth" ][ "day" ] ) {
		let birth_day_element = document.getElementById( "user_birth_day" );
		birth_day_element.value = user_info[ "identity" ][ "date_of_birth" ][ "day" ];
	}
	if ( user_info[ "identity" ][ "date_of_birth" ][ "month" ] ) {
		let birth_month_element = document.getElementById( "user_birth_month" );
		birth_month_element.value = user_info[ "identity" ][ "date_of_birth" ][ "month" ];
	}
	if ( user_info[ "identity" ][ "date_of_birth" ][ "year" ] ) {
		let birth_year_element = document.getElementById( "user_birth_year" );
		birth_year_element.value = user_info[ "identity" ][ "date_of_birth" ][ "year" ];
	}

	// Update Dynamic Stuff

	if ( user_info[ "verified" ] ) {
		$( "#verified-img" ).show();
		$( "#verified-button-text" ).text( "" );
	}

	if ( user_info[ "family_members" ] ) {
		for ( let i = 0; i < user_info[ "family_members" ].length; ++i ) {
			let family_member_ulid = on_add_family_member();
			window.FAMILY_MEMBERS[ family_member_ulid ] = user_info[ "family_members" ][ i ];

			let age = false;
			let male = document.getElementById( `user_family_member_${family_member_ulid}_gender_male` );
			let male_text = false;
			let female = document.getElementById( `user_family_member_${family_member_ulid}_gender_female` );
			let female_text = false;
			let spouse = document.getElementById( `user_family_member_${family_member_ulid}_spouse` );

			if ( user_info[ "family_members" ][ i ].age ) {
				age = document.getElementById( `user_family_member_${family_member_ulid}_age` );
				age.value = user_info[ "family_members" ][ i ].age;
			}

			if ( user_info[ "family_members" ][ i ].sex ) {
				if ( user_info[ "family_members" ][ i ].sex === "male" ) {
					male.checked = true;
				} else {
					female.checked = true;
				}
			}

			if ( user_info[ "family_members" ][ i ].age < 18 ) {
				male_text = document.getElementById( `user_family_member_${family_member_ulid}_gender_label_male` );
				male_text.textContent = "Boy";
				female_text = document.getElementById( `user_family_member_${family_member_ulid}_gender_label_female` );
				female_text.textContent = "Girl";

				spouse.parentNode.parentNode.style.display = "none";
			}

			if ( user_info[ "family_members" ][ i ].spouse ) {
				spouse.checked = true;
			}

		}
	}

	if ( user_info[ "barcodes" ] ) {
		console.log( "barcodes" , user_info[ "barcodes" ] );
		for ( let i = 0; i < user_info[ "barcodes" ].length; ++i ) {
			let barcode_ulid = on_add_barcode(); // add barcode to DOM
			let barcode_id = `user_barcode_${barcode_ulid}`;
			let barcode_input_element = document.getElementById( barcode_id );
			barcode_input_element.value = user_info[ "barcodes" ][ i ];
			window.BARCODES[ barcode_ulid ] = user_info[ "barcodes" ][ i ];
		}
	}

	if ( user_info[ "spanish" ] ) {
		document.getElementById( "user_spanish" ).checked = user_info[ "spanish" ];
	}

}