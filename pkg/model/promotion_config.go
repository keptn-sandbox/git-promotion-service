package model

type PromotionConfig struct {
	APIVersion *string             `yaml:"apiVersion"`
	Kind       *string             `yaml:"kind"`
	Spec       PromotionConfigSpec `yaml:"spec"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

type PromotionConfigSpec struct {
	Strategy *string `yaml:"strategy"`
	Target   Target  `yaml:"target"`
	Paths    []Path  `yaml:"paths"`
}

type Target struct {
	Repo     *string `yaml:"repo"`
	Secret   *string `yaml:"secret"`
	Provider *string `yaml:"provider"`
}

type Path struct {
	Source *string `yaml:"source"`
	Target *string `yaml:"target"`
}
