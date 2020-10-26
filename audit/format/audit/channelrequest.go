package audit

type PayloadChannelRequestUnknownType struct {
	RequestType string `json:"requestType" yaml:"requestType"`
}

type PayloadChannelRequestDecodeFailed struct {
	RequestType string `json:"requestType" yaml:"requestType"`
	Reason      string `json:"reason" yaml:"reason"`
}

type PayloadChannelRequestFailed struct {
	RequestType string `json:"requestType" yaml:"requestType"`
	Reason      string `json:"reason" yaml:"reason"`
}

type PayloadChannelRequestSetEnv struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type PayloadChannelRequestExec struct {
	Program string `json:"program" yaml:"program"`
}

type PayloadChannelRequestPty struct {
	Columns uint `json:"columns" yaml:"columns"`
	Rows    uint `json:"rows" yaml:"rows"`
}

type PayloadChannelRequestShell struct {
}

type PayloadChannelRequestSignal struct {
	Signal string `json:"signal" yaml:"signal"`
}

type PayloadChannelRequestSubsystem struct {
	Subsystem string `json:"subsystem" yaml:"subsystem"`
}

type PayloadChannelRequestWindow struct {
	Columns uint `json:"columns" yaml:"columns"`
	Rows    uint `json:"rows" yaml:"rows"`
}
