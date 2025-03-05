package versionfile

const InstalledVersionStructVersion = 20250303

type InstalledVersion struct {
	Name               string              `json:"name"`
	Version            string              `json:"version"`
	ImageDigest        string              `json:"image_digest,omitempty"`
	BinaryDigest       string              `json:"binary_digest,omitempty"`
	BinaryArchitecture string              `json:"binary_arch,omitempty"`
	InstalledFrom      string              `json:"installed_from,omitempty"`
	LastCheckedDate    string              `json:"last_checked_date,omitempty"`
	InstallDate        string              `json:"install_date,omitempty"`
	StructVersion      int64               `json:"struct_version"`
	Metadata           map[string][]string `json:"metadata,omitempty"`
}

func EmptyInstalledVersion() *InstalledVersion {
	i := new(InstalledVersion)
	i.StructVersion = InstalledVersionStructVersion
	return i
}

// Equal compares the `Name` and `BinaryDigest`
func (f *InstalledVersion) Equal(other *InstalledVersion) bool {
	return f.Name == other.Name && f.BinaryDigest == other.BinaryDigest
}
