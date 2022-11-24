package provider

type IConfigProvider interface {
	LookUpEnv(variable string) (string, bool)
}

type NopConfigProvider struct {
	IConfigProvider
}

func (n *NopConfigProvider) LookUpEnv(variable string) (string, bool) {
	return "", false
}

var _ IConfigProvider = &NopConfigProvider{}
