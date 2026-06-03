package viva_api

// Ключи шаблонов уведомлений каталога (catalog/safekid.json → notifications.*).
// Номера СМС — по таблице сообщений оператора (№2–№13).
const (
	NotifyWelcomeTrial       = "welcome_trial"       // СМС №2
	NotifyLicense            = "license"             // СМС №3
	NotifyWelcomePaid        = "welcome_paid"        // СМС №4
	NotifyTrialExpires       = "trial_expires"       // СМС №5
	NotifyServiceDeactivated = "service_deactivated" // СМС №6
	NotifyAlreadyActive      = "already_active"      // СМС №8
	NotifyAlreadyDeactivated = "already_deactivated" // СМС №9
	NotifyLanguageChanged    = "language_changed"    // СМС №10–№12
	NotifyUnknownCommand     = "unknown_command"     // СМС №13
	NotifyOtpLanding         = "otp_landing"         // СМС со ссылкой на лендинг после MO «1» (текст уточняет оператор)
)
