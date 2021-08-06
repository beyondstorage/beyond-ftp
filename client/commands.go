package client

const (
	USER = "USER"
	PASS = "PASS"
	AUTH = "AUTH"
	PROT = "PROT"
	PBSZ = "PBSZ"
	FEAT = "FEAT"
	SYST = "SYST"
	NOOP = "NOOP"
	OPTS = "OPTS"
	ABOR = "ABOR"
	SIZE = "SIZE"
	STAT = "STAT"
	MDTM = "MDTM"
	RETR = "RETR"
	STOR = "STOR"
	APPE = "APPE"
	DELE = "DELE"
	RNFR = "RNFR"
	RNTO = "RNTO"
	ALLO = "ALLO"
	REST = "REST"
	SITE = "SITE"
	CWD  = "CWD"
	PWD  = "PWD"
	CDUP = "CDUP"
	NLST = "NLST"
	LIST = "LIST"
	MKD  = "MKD"
	RMD  = "RMD"
	XMKD = "XMKD"
	XRMD = "XRMD"
	XPWD = "XPWD"
	XCUP = "XCUP"
	TYPE = "TYPE"
	PASV = "PASV"
	EPSV = "EPSV"
	PORT = "PORT"
	QUIT = "QUIT"
	ACCT = "ACCT"
	ADAT = "ADAT"
	CCC  = "CCC"
	CONF = "CONF"
	ENC  = "ENC"
	EPRT = "EPRT"
	HELP = "HELP"
	LANG = "LANG"
	MIC  = "MIC"
	MLSD = "MLSD"
	MLST = "MLST"
	MODE = "MODE"
	REIN = "REIN"
	SMNT = "SMNT"
	STOU = "STOU"
	STRU = "STRU"
)

// CommandDescription defines which function should be used and if it should be
// open to anyone or only logged in users.
type CommandDescription struct {
	Open bool           // Open to clients without auth.
	Fn   func(*Handler) // Function to handle it.
}

var commandsMap map[string]*CommandDescription

func init() {
	// This is shared between FTPServer instances as there's no point in making
	// the FTP commands behave differently between them.
	// Whole commands can be found here: https://tools.ietf.org/html/rfc5797

	commandsMap = make(map[string]*CommandDescription)

	// Authentication.
	commandsMap[USER] = &CommandDescription{Fn: (*Handler).handleUSER, Open: true}
	commandsMap[PASS] = &CommandDescription{Fn: (*Handler).handlePASS, Open: true}

	// TLS handling.
	commandsMap[AUTH] = nil
	commandsMap[PROT] = nil
	commandsMap[PBSZ] = nil

	// Misc.
	commandsMap[FEAT] = &CommandDescription{Fn: (*Handler).handleFEAT, Open: true}
	commandsMap[SYST] = &CommandDescription{Fn: (*Handler).handleSYST, Open: true}
	commandsMap[NOOP] = &CommandDescription{Fn: (*Handler).handleNOOP, Open: true}
	commandsMap[OPTS] = &CommandDescription{Fn: (*Handler).handleOPTS, Open: true}
	commandsMap[ABOR] = &CommandDescription{Fn: (*Handler).handleABOR}

	// File access.
	commandsMap[SIZE] = &CommandDescription{Fn: (*Handler).handleSIZE}
	commandsMap[STAT] = &CommandDescription{Fn: (*Handler).handleSTAT}
	commandsMap[MDTM] = &CommandDescription{Fn: (*Handler).handleMDTM}
	commandsMap[RETR] = &CommandDescription{Fn: (*Handler).handleRETR}
	commandsMap[STOR] = &CommandDescription{Fn: (*Handler).handleSTOR}
	commandsMap[APPE] = &CommandDescription{Fn: (*Handler).handleAPPE}
	commandsMap[DELE] = &CommandDescription{Fn: (*Handler).handleDELE}
	commandsMap[RNFR] = &CommandDescription{Fn: (*Handler).handleRNFR}
	commandsMap[RNTO] = &CommandDescription{Fn: (*Handler).handleRNTO}
	commandsMap[ALLO] = &CommandDescription{Fn: (*Handler).handleALLO}
	commandsMap[REST] = &CommandDescription{Fn: (*Handler).handleREST}
	commandsMap[SITE] = nil

	// Directory handling.
	commandsMap[CWD] = &CommandDescription{Fn: (*Handler).handleCWD}
	commandsMap[PWD] = &CommandDescription{Fn: (*Handler).handlePWD}
	commandsMap[CDUP] = &CommandDescription{Fn: (*Handler).handleCDUP}
	commandsMap[NLST] = &CommandDescription{Fn: (*Handler).handleLIST}
	commandsMap[LIST] = &CommandDescription{Fn: (*Handler).handleLIST}
	commandsMap[MKD] = &CommandDescription{Fn: (*Handler).handleMKD}
	commandsMap[RMD] = &CommandDescription{Fn: (*Handler).handleRMD}

	// XMKD, XRMD, XPWD, XCUP
	// Implementation note:  Deployed FTP clients still make use of the
	// deprecated commands and most FTP servers support them as aliases
	// for the standard commands.
	// ref: https://tools.ietf.org/html/rfc5797
	commandsMap[XMKD] = &CommandDescription{Fn: (*Handler).handleMKD}
	commandsMap[XRMD] = &CommandDescription{Fn: (*Handler).handleRMD}
	commandsMap[XPWD] = &CommandDescription{Fn: (*Handler).handlePWD}
	commandsMap[XCUP] = &CommandDescription{Fn: (*Handler).handleCWD}

	// Connection handling.
	commandsMap[TYPE] = &CommandDescription{Fn: (*Handler).handleTYPE}
	commandsMap[PASV] = &CommandDescription{Fn: (*Handler).handlePASV}
	commandsMap[EPSV] = &CommandDescription{Fn: (*Handler).handlePASV}
	commandsMap[PORT] = &CommandDescription{Fn: (*Handler).handlePORT}
	commandsMap[QUIT] = &CommandDescription{Fn: (*Handler).handleQUIT, Open: true}

	// Not Supported command.
	commandsMap[ACCT] = nil
	commandsMap[ADAT] = nil
	commandsMap[CCC] = nil
	commandsMap[CONF] = nil
	commandsMap[ENC] = nil
	commandsMap[EPRT] = nil
	commandsMap[HELP] = nil
	commandsMap[LANG] = nil
	commandsMap[MIC] = nil
	commandsMap[MLSD] = nil
	commandsMap[MLST] = nil
	commandsMap[MODE] = nil
	commandsMap[REIN] = nil
	commandsMap[SMNT] = nil
	commandsMap[STOU] = nil
	commandsMap[STRU] = nil
}
