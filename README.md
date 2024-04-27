# Master's Closet Tracking Server

## Onboarding Experience
1. Admin Enters Provided First and Last Name
2. Server Redirects to `/admin/user/new/handoff/${new-users-uuid}`
3. New user scans Hand-Off QR code with their phone
4. Scanned QR Hand-Off Code takes them to a silent login page that stores a permanent login cookie.
	- `/user/login/fresh/${new-users-uuid}`

---

## To Re-Enter
1. They scan a QR code on a poster at the front door or just go to `/checkin`
2. If they have a cookie stored it redirects to `/user/checkin/display/${uuid}`
3. Admin Scans and checks-in/validates their QR-Code with stored uuid

---

## TODO
1. Make config editable via html
	- gather misc stuff like print speed
2. cache stuff in production :
	- https://docs.gofiber.io/api/middleware/cache
3. Just let a barcode check-in a user. Avoids an extra call
	- GET /admin/user/get/barcode/:barcode
	- GET /admin/user/checkin/test/:uuid
	- GET /admin/user/checkin/:uuid
4. Add Admin Manual Override Routes
	- Override Check-In Too Soon
	- User forgot phone
	- User has new phone
	- option to text hand-off link if user can't scan qrcode for some reason
5. Fix User Fields :
	- Authorized Aliases
6. Fix Docker
7. Use time functions
	- `time.Now().After(lastFetched.Add(CachePeriod))` ?
8. Change "usernames" DB bucket for key=${uuid}_username , value=Username
	- keeps only uuids as keys
9. audio ml stuff for new user creation and check-in
10. Fix ui.js#793
	- `document.getElementById( barcode_id ).focus();`
	- make this optional , so that the edit page doesn't use this
11. Fix Username/NameString to be Title Case?
12. Make family-member management more streamlined

## Misc

- https://offnova.com/pages/download
- `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`
- `brew install cups`
- Windows 11 S-Mode ??
	- https://support.microsoft.com/en-us/windows/switching-out-of-s-mode-in-windows-4f56d9be-99ec-6983-119f-031bfb28a307
	- `ms-windows-store://pdp/?productid=BF712690PMLF&OCID=windowssmodesupportpage`

- https://github.com/apple/cups/releases
- `git clone https://github.com/apple/cups`
- `cd cups`
- `./configure --prefix="$(pwd)/build"`
- `./configure --prefix="/Applications/MCT.app/Contents/Resources"`
- `make`
- `sudo make install`

- `sudo rsync -av /usr/local/Cellar/cups/$(brew list --versions cups | awk '{print $2}') ./cups`

## Time Zone Data for Windows

- https://stackoverflow.com/questions/48439363/missing-location-in-call-to-time-in