package types

import (
	// fiber "github.com/gofiber/fiber/v2"
)

type RedisConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	DB int "json:db"
	Password string "json:password"
}

type BalanceGeneral struct {
	Total int `json:"total"`
	Tops int `json:"tops"`
	Bottoms int `json:"bottoms"`
	Dresses int `json:"dresses"`
}

type BalanceConfig struct {
	General BalanceGeneral `json:"general"`
	Shoes int `json:"shoes"`
	Seasonals int `json:"seasonals"`
	Accessories int `json:"accessories"`
}

type PrinterConfig struct {
	PageWidth float64 `json:"page_width"`
	PageHeight float64 `json:"page_height"`
	FontName string `json:"font_name"`
	FontPath string `json:"font_path"`
	PrinterName string `json:"printer_name"`
	LogoFilePath string `json:"logo_file_path"`
}

type EmailConfig struct {
	SMTPServer string `json:"smtp_server"`
	SMTPServerUrl string `json:"smtp_server_url"`
	SMTPAuthEmail string `json:"smtp_auth_email"`
	SMTPAuthPassword string `json:"smtp_auth_password"`
	From string `json:"from"`
}

type ConfigFile struct {
	ServerBaseUrl string `json:"server_base_url"`
	ServerPort string `json:"server_port"`
	ServerAPIKey string `json:"server_api_key"`
	ServerCookieSecret string `json:"server_cookie_secret"`
	ServerCookieAdminSecretMessage string `json:"server_cookie_admin_secret_message"`
	ServerCookieSecretMessage string `json:"server_cookie_secret_message"`
	ServerLiveUrl string `json:"server_live_url"`
	LocalHostUrl string `json:"local_host_url"`
	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
	TimeZone string `json:"time_zone"`
	BoltDBPath string `json:"bolt_db_path"`
	BoltDBEncryptionKey string `json:"bolt_db_encryption_key"`
	BleveSearchPath string `json:"bleve_search_path"`
	CheckInCoolOffDays int `json:"check_in_cooloff_days"`
	TwilioClientID string `json:"twilio_client_id"`
	TwilioAuthToken string `json:"twilio_auth_token"`
	TwilioSMSFromNumber string `json:"twilio_sms_from_number"`
	Email EmailConfig `json:"email"`
	Redis RedisConfig `json:"redis"`
	IPBlacklist []string `json:"ip_blacklist"`
	IPInfoToken string `json:"ip_info_token"`
	Balance BalanceConfig `json:"balance"`
	Printer PrinterConfig `json:"printer"`
	FingerPrint string `json:"finger_print"`
}

type AListResponse struct {
	UUIDS []string `json:"uuids"`
}

type RedisMultiCommand struct {
	Command string `json:"type"`
	Key string `json:"key"`
	Args string `json:"args"`
}

type SearchItem struct {
	UUID string
	Name string
}