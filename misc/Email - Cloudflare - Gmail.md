<center><h1>Email : Cloudflare <-> Gmail</h1></center>

1. Cloudflare Enable Email on Domain
	- Goto "Websites"
	- Click into a site
	- Click "Email" on the left
	- It should prompt you to "Add records and enable"
2. Cloudflare Add Email Route
	- Go back into the website settings
	- Click "Email Routing"
	- Click "Create address"
	- Add custom address
		- Action = "Send to an email"
		- Destination = your gmail address
3. Cloudflare Enable Gmail Forwarding
	- Go back into the website settings
	- Click "DNS" on the left
	- Edit the "TXT" record
		- `v=spf1 include:_spf.mx.cloudflare.net include:_spf.google.com ~all`
4. Enable 2FA on Google Account
	- https://www.google.com/landing/2step
	- https://myaccount.google.com/signinoptions/two-step-verification
5. Create an "app" password
	- https://security.google.com/settings/security/apppasswords
	- Select App --> Other (Custom name)
	- Name it whatever
	- Press "Generate"
	- Copy the password to config.json
6. Create alias in Gmail
	- In Gmail , goto : Settings → Accounts and Import → Send mail as :
		- https://mail.google.com/mail/u/0/#settings/general
		- Click "Add another email address"
			- allow popups
		- Name = blank , or whatever
		- Email address = cloudflare email address
		- Un-check "Treat as an alias" !!!
		- Click "Next Step >>"
		- SMTP Server  = smtp.gmail.com
		- SMTP Server Port = 587
		- Username = your gmail username
		- Password = "app" password you created
		- Select "Secured connection using TLS"
		- Click "Add Account"
		- Wait on email for conformation code , or just click the link in the email

---

- https://community.cloudflare.com/t/solved-how-to-use-gmail-smtp-to-send-from-an-email-address-which-uses-cloudflare-email-routing/382769/2
- https://community.cloudflare.com/t/adding-dns-records/52718