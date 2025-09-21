package consts

const (
	BusCodeUserStartCode = (iota+1)*1000 + 1
	BusCodeSettingStartCode
	BusCodeUserGroupStartCode
	BusCodeStorageStartCode
	BusCodeCloudTokenStartCode
	BusCodeFileStartCode
	BusCodeTaskStateStartCode
	BusCodeStorageAdvanceStartCode

	BusCodeMiddlewareAuth = 99100 + 1
)
