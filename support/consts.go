package support

type Integration struct {
	DataSource string
	Url        string
}

type Parameters struct {
	DryRun    bool
	Slow      bool
	Custom    string
	InputFile string
}

const (
	DCM_EDT string = "DCM_EDT"
	DBM_EDT        = "DBM_EDT"
	SFTP           = "SFTP"
	S3             = "S3"
	TDD_V5         = "TDD_V5"
	AD_FORM        = "AD_FORM"
	FTK_EDT        = "FTK_EDT"
	SZK_EDT        = "SZK_EDT"
	CUSTOM         = "CUSTOM"
)

var IntegrationMappings = map[string]Integration{
	DCM_EDT: {DataSource: DCM_EDT, Url: "http://edtsclr1-1-prod-dca.agkn.net:8080/dcm2/task/runnow/%s?force=true"},
	DBM_EDT: {DataSource: DBM_EDT, Url: "http://edtsclr1-1-prod-dca.agkn.net:8080/dbm2/task/runnow/%s?force=true"},
	FTK_EDT: {DataSource: DBM_EDT, Url: "http://edtsclr1-1-prod-dca.agkn.net:8080/flashtalking/task/runnow/%s?force=true"},
	SFTP:    {DataSource: SFTP, Url: "http://edtsitesclr1-5-qa-dca.agkn.net:8080/sftp/task/runnow/%s?force=true"},
	S3:      {DataSource: S3, Url: "http://edtsclr1-1-prod-dca.agkn.net:8080/s3/task/runnow/%s?force=true"},
	TDD_V5:  {DataSource: TDD_V5, Url: "http://edttrddsk1-1-prod-dca.agkn.net:8080/tradedesk-v5/task/runnow/%s?force=true"},
	AD_FORM: {DataSource: AD_FORM, Url: "http://edtsclr1-1-prod-dca.agkn.net:8080/adform/task/runnow/%s?force=true"},
	SZK_EDT	: {DataSource: SZK_EDT, Url: "http://edtsclr1-1-prod-dca.agkn.net:8080/sizmek/task/runnow/%s?force=true"},
}

const DefaultInputFile = "./config/ids.txt"
