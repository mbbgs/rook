package consts

const (
    
    AGREEMENT_FILE  = "agreement.rook"
    
    APP_INIT        = "app:init"
	  APP_BOOT        = "app:boot"
	  APP_READY     	= "app:ready"
	  APP_SHUTDOWN	  = "app:shutdown"
	  APP_TERMINATE		= "app:terminate"
	  DROP_TABLE      = "app:drop"
	  
  	USER_LOGIN      = "user:login"
  	USER_LOGOUT     = "user:logout"
  	USER_LOGGED_IN  = "user:logged in"
  	USER_REGISTRATION = "user: registration "
  	RESET_PASSWORD  =  "user:reset password"
  	
  	SECRET_ROOK     = ".secret.rook"
  	STORE_FILE_PATH = "storook"
  	ATTEMPTS_PATH   = ".attempts.rook"
  	ROOK_LOG        = ".log.rook"
  
  	SALT_SIZE       = 16
  	MAX_ATTEMPTS    = 06
  	MAX_M_ATTEMPTS  = 05
  	
  	TERMS           = `
  	ROOK SECURITY POLICY & TERMS

1. No Cloud Storage – All data is local.
2. Zero Trust – No implicit trust.
3. 5 Failed Attempts – Auto-wipe triggered.
4. No Recovery – Use at your own risk.
5. Immediate Termination – No support, no mercy.
`
)
