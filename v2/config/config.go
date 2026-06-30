// Package config defines the typed application configuration and a thread-safe
// Manager that loads/persists it from BoltDB. In v1 config was parsed from JSON
// once at startup and never written back (the only "writes" rewrote .js source
// files by hardcoded line number). v2 keeps the live config in the db and edits
// it through the settings panel.
package config

// EmailConfig holds SMTP credentials for outbound mail.
type EmailConfig struct {
	SMTPServer       string `json:"smtp_server"`
	SMTPServerURL    string `json:"smtp_server_url"`
	SMTPAuthEmail    string `json:"smtp_auth_email"`
	SMTPAuthPassword string `json:"smtp_auth_password"`
	From             string `json:"from"`
}

// RedisConfig holds the optional Redis connection. (v1 had malformed struct
// tags here: `json:db` / `json:password` — fixed below.)
type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	DB       int    `json:"db"`
	Password string `json:"password"`
}

// BalanceGeneral is the per-family-member allowance for general clothing.
type BalanceGeneral struct {
	Total   int `json:"total"`
	Tops    int `json:"tops"`
	Bottoms int `json:"bottoms"`
	Dresses int `json:"dresses"`
}

// BalanceConfig is the per-checkin-period clothing allowance.
type BalanceConfig struct {
	General     BalanceGeneral `json:"general"`
	Shoes       int            `json:"shoes"`
	Seasonals   int            `json:"seasonals"`
	Accessories int            `json:"accessories"`
}

// PrinterConfig describes the barcode label printer.
type PrinterConfig struct {
	Speed        int     `json:"speed"`
	PageWidth    float64 `json:"page_width"`
	PageHeight   float64 `json:"page_height"`
	FontName     string  `json:"font_name"`
	FontPath     string  `json:"font_path"`
	PrinterName  string  `json:"printer_name"`
	LogoFilePath string  `json:"logo_file_path"`
}

// Config is the whole application configuration. Field tags match v1 so existing
// seed files load unchanged.
type Config struct {
	ServerBaseURL                  string `json:"server_base_url"`
	ServerPort                     string `json:"server_port"`
	ServerAPIKey                   string `json:"server_api_key"`
	ServerCookieSecret             string `json:"server_cookie_secret"`
	ServerCookieAdminSecretMessage string `json:"server_cookie_admin_secret_message"`
	ServerCookieSecretMessage      string `json:"server_cookie_secret_message"`
	ServerLiveURL                  string `json:"server_live_url"`
	LocalHostURL                   string `json:"local_host_url"`
	RemoteHostURL                  string `json:"remote_host_url"`
	RemoteHostAPIKey               string `json:"remote_host_api_key"`
	RemoteHostHeaderPrefix         string `json:"remote_host_header_prefix"`
	RemoteHostClientID             string `json:"remote_host_client_id"`
	AdminUsername                  string `json:"admin_username"`
	AdminPassword                  string `json:"admin_password"`
	TimeZone                       string `json:"time_zone"`
	ServeLocation                  string `json:"serve_location"`
	BoltDBPath                     string `json:"bolt_db_path"`
	BoltDBEncryptionKey            string `json:"bolt_db_encryption_key"`
	BleveSearchPath                string `json:"bleve_search_path"`
	LevenshteinDistanceThreshold   int    `json:"levenshtein_distance_threshold"`
	CheckInCoolOffDays             int    `json:"check_in_cooloff_days"`
	AutoVerifyZipcode              string `json:"auto_verify_zipcode"`
	TwilioClientID                 string `json:"twilio_client_id"`
	TwilioAuthToken                string `json:"twilio_auth_token"`
	TwilioSMSFromNumber            string `json:"twilio_sms_from_number"`
	OpenAIAPIKey                   string `json:"open_ai_api_key"`

	Email       EmailConfig   `json:"email"`
	Redis       RedisConfig   `json:"redis"`
	IPBlacklist []string      `json:"ip_blacklist"`
	IPInfoToken string        `json:"ip_info_token"`
	Balance     BalanceConfig `json:"balance"`
	Printer     PrinterConfig `json:"printer"`
	FingerPrint string        `json:"finger_print"`
}

// FieldMeta drives the settings panel: how to label, group, mask, and gate a
// configuration field. Path is the dotted json path ("email.smtp_auth_password").
type FieldMeta struct {
	Path            string `json:"path"`
	Label           string `json:"label"`
	Section         string `json:"section"`
	Kind            string `json:"kind"` // string | int | float | bool | list
	Secret          bool   `json:"secret"`
	RestartRequired bool   `json:"restart_required"`
	ReadOnly        bool   `json:"read_only"` // bootstrap-only fields
	Help            string `json:"help,omitempty"`
}

// Fields returns the ordered metadata the settings panel renders. Fields not
// listed here are persisted but hidden from the editor (e.g. finger_print).
func Fields() []FieldMeta {
	return []FieldMeta{
		// Server / bind — changing these needs a restart.
		{Path: "server_base_url", Label: "Server Base URL", Section: "Server", Kind: "string", RestartRequired: true},
		{Path: "server_live_url", Label: "Server Live URL", Section: "Server", Kind: "string", RestartRequired: true},
		{Path: "local_host_url", Label: "Local Host URL", Section: "Server", Kind: "string"},
		{Path: "server_port", Label: "Server Port", Section: "Server", Kind: "string", RestartRequired: true, Help: "Bind port; takes effect after restart."},
		{Path: "serve_location", Label: "Serve Location", Section: "Server", Kind: "string", Help: "\"local\" or \"remote\""},
		{Path: "time_zone", Label: "Time Zone", Section: "Server", Kind: "string", RestartRequired: true},

		// Security / secrets.
		{Path: "server_api_key", Label: "Server API Key", Section: "Security", Kind: "string", Secret: true},
		{Path: "server_cookie_secret", Label: "Cookie Secret", Section: "Security", Kind: "string", Secret: true, RestartRequired: true},
		{Path: "server_cookie_admin_secret_message", Label: "Admin Cookie Message", Section: "Security", Kind: "string", Secret: true},
		{Path: "server_cookie_secret_message", Label: "User Cookie Message", Section: "Security", Kind: "string", Secret: true},
		{Path: "admin_username", Label: "Admin Username", Section: "Security", Kind: "string", Secret: true},
		{Path: "admin_password", Label: "Admin Password", Section: "Security", Kind: "string", Secret: true},
		{Path: "bolt_db_path", Label: "BoltDB Path", Section: "Security", Kind: "string", ReadOnly: true, Help: "Bootstrap-only; change via seed file + restart."},
		{Path: "bolt_db_encryption_key", Label: "BoltDB Encryption Key", Section: "Security", Kind: "string", Secret: true, ReadOnly: true, Help: "Bootstrap-only; changing it would orphan existing records."},
		{Path: "bleve_search_path", Label: "Bleve Search Path", Section: "Security", Kind: "string", ReadOnly: true},

		// Check-in behavior — safe to apply live.
		{Path: "check_in_cooloff_days", Label: "Check-In Cooloff (days)", Section: "Check-In", Kind: "int"},
		{Path: "levenshtein_distance_threshold", Label: "Similarity Threshold", Section: "Check-In", Kind: "int"},
		{Path: "auto_verify_zipcode", Label: "Auto-Verify Zipcode", Section: "Check-In", Kind: "string", Help: "Users with this zipcode are auto-marked Verified. Default 45424."},

		// Balance allowances — safe to apply live.
		{Path: "balance.general.total", Label: "General Total", Section: "Balance", Kind: "int"},
		{Path: "balance.general.tops", Label: "Tops", Section: "Balance", Kind: "int"},
		{Path: "balance.general.bottoms", Label: "Bottoms", Section: "Balance", Kind: "int"},
		{Path: "balance.general.dresses", Label: "Dresses", Section: "Balance", Kind: "int"},
		{Path: "balance.shoes", Label: "Shoes", Section: "Balance", Kind: "int"},
		{Path: "balance.seasonals", Label: "Seasonals", Section: "Balance", Kind: "int"},
		{Path: "balance.accessories", Label: "Accessories", Section: "Balance", Kind: "int"},

		// Email.
		{Path: "email.smtp_server", Label: "SMTP Server", Section: "Email", Kind: "string"},
		{Path: "email.smtp_server_url", Label: "SMTP Server URL", Section: "Email", Kind: "string"},
		{Path: "email.smtp_auth_email", Label: "SMTP Auth Email", Section: "Email", Kind: "string"},
		{Path: "email.smtp_auth_password", Label: "SMTP Auth Password", Section: "Email", Kind: "string", Secret: true},
		{Path: "email.from", Label: "From Address", Section: "Email", Kind: "string"},

		// Twilio.
		{Path: "twilio_client_id", Label: "Twilio Client ID", Section: "Twilio", Kind: "string", Secret: true},
		{Path: "twilio_auth_token", Label: "Twilio Auth Token", Section: "Twilio", Kind: "string", Secret: true},
		{Path: "twilio_sms_from_number", Label: "Twilio From Number", Section: "Twilio", Kind: "string"},

		// Integrations.
		{Path: "open_ai_api_key", Label: "OpenAI API Key", Section: "Integrations", Kind: "string", Secret: true},
		{Path: "ip_info_token", Label: "IPInfo Token", Section: "Integrations", Kind: "string", Secret: true},

		// Remote sync (deferred functionally, but configurable).
		{Path: "remote_host_url", Label: "Remote Host URL", Section: "Remote Sync", Kind: "string"},
		{Path: "remote_host_api_key", Label: "Remote Host API Key", Section: "Remote Sync", Kind: "string", Secret: true},
		{Path: "remote_host_header_prefix", Label: "Remote Header Prefix", Section: "Remote Sync", Kind: "string"},
		{Path: "remote_host_client_id", Label: "Remote Client ID", Section: "Remote Sync", Kind: "string", Secret: true},

		// Printer.
		{Path: "printer.speed", Label: "Print Speed", Section: "Printer", Kind: "int"},
		{Path: "printer.page_width", Label: "Page Width", Section: "Printer", Kind: "float"},
		{Path: "printer.page_height", Label: "Page Height", Section: "Printer", Kind: "float"},
		{Path: "printer.font_name", Label: "Font Name", Section: "Printer", Kind: "string"},
		{Path: "printer.printer_name", Label: "Printer Name", Section: "Printer", Kind: "string"},

		// Redis.
		{Path: "redis.host", Label: "Redis Host", Section: "Redis", Kind: "string"},
		{Path: "redis.port", Label: "Redis Port", Section: "Redis", Kind: "string"},
		{Path: "redis.db", Label: "Redis DB", Section: "Redis", Kind: "int"},
		{Path: "redis.password", Label: "Redis Password", Section: "Redis", Kind: "string", Secret: true},
	}
}

// metaByPath indexes Fields() for quick lookup during validation/diff.
func metaByPath() map[string]FieldMeta {
	out := map[string]FieldMeta{}
	for _, f := range Fields() {
		out[f.Path] = f
	}
	return out
}
