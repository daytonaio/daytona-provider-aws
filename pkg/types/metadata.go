package types

type TargetMetadata struct {
	InstanceId string
	Tags       map[string]string
	IsRunning  bool
	Created    string
}
