package types

type WorkspaceMetadata struct {
	InstanceId string
	Tags       map[string]string
	IsRunning  bool
	Created    string
}
