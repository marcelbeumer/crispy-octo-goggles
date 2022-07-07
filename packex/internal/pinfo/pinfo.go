package pinfo

type Package struct {
	Name string
}

func (pkg *Package) Resolve() error {
	return nil
}

func (pkg *Package) String() string {
	return ""
}

func New(nameOrPath string) *Package {
	return &Package{Name: ""}
}
