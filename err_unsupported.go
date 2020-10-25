package yamlreg

type ErrUnsupportedYAMLFeature struct {
	Feature string
}

func (f ErrUnsupportedYAMLFeature) Error() string {
	return "unsupported YAML feature: " + f.Feature
}
