package modconfig

import (
	"github.com/Masterminds/semver/v3"
	"testing"
)

func TestDependencyVersion_Equal(t *testing.T) {

	tests := []struct {
		name string
		l    *DependencyVersion
		r    *DependencyVersion
		want bool
	}{
		{
			name: "nil other",
			l: &DependencyVersion{
				Version: semver.MustParse("1.0.0"),
			},
			want: false,
		},
		{
			name: "both have same version",
			l: &DependencyVersion{
				Version: semver.MustParse("1.0.0"),
			},
			r: &DependencyVersion{
				Version: semver.MustParse("1.0.0"),
			},
			want: true,
		},
		{
			name: "both have different version",
			l: &DependencyVersion{
				Version: semver.MustParse("1.0.0"),
			},
			r: &DependencyVersion{
				Version: semver.MustParse("2.0.0"),
			},
			want: false,
		},
		{
			name: "both have same branch",
			l: &DependencyVersion{
				Branch: "master",
			},
			r: &DependencyVersion{
				Branch: "master",
			},
			want: true,
		},
		{
			name: "both have different branch",
			l: &DependencyVersion{
				Branch: "master",
			},
			r: &DependencyVersion{
				Branch: "develop",
			},
			want: false,
		},
		{
			name: "both have same filepath",
			l: &DependencyVersion{
				FilePath: "path/to/mod",
			},
			r: &DependencyVersion{
				FilePath: "path/to/mod",
			},
			want: true,
		},
		{
			name: "both have different filepath",
			l: &DependencyVersion{
				FilePath: "path/to/mod",
			},
			r: &DependencyVersion{
				FilePath: "path/to/othermod",
			},
			want: false,
		},
		{
			name: "both have same tag",
			l: &DependencyVersion{
				Tag: "v1.0.0",
			},
			r: &DependencyVersion{
				Tag: "v1.0.0",
			},
			want: true,
		},
		{
			name: "both have different tag",
			l: &DependencyVersion{
				Tag: "v1.0.0",
			},
			r: &DependencyVersion{
				Tag: "v2.0.0",
			},
			want: false,
		},

		{
			name: "version and branch",
			l: &DependencyVersion{
				Version: semver.MustParse("1.0.0"),
			},
			r: &DependencyVersion{
				Branch: "master",
			},
			want: false,
		},
		{
			name: "version and filepath",
			l: &DependencyVersion{
				Version: semver.MustParse("1.0.0"),
			},
			r: &DependencyVersion{
				FilePath: "path/to/mod",
			},
			want: false,
		},
		{
			name: "version and tag",
			l: &DependencyVersion{
				Version: semver.MustParse("1.0.0"),
			},
			r: &DependencyVersion{
				Tag: "v1.0.0",
			},
			want: false,
		},
		{
			name: "branch and filepath",
			l: &DependencyVersion{
				Branch: "master",
			},
			r: &DependencyVersion{
				FilePath: "path/to/mod",
			},
			want: false,
		},
		{
			name: "branch and tag",
			l: &DependencyVersion{
				Branch: "master",
			},
			r: &DependencyVersion{
				Tag: "v1.0.0",
			},
			want: false,
		},
		{
			name: "filepath and tag",
			l: &DependencyVersion{
				FilePath: "path/to/mod",
			},
			r: &DependencyVersion{
				Tag: "v1.0.0",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.Equal(tt.r); got != tt.want {
				t.Errorf("DependencyVersion.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
