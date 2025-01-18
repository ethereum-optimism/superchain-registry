package config

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestHardforkTime_MarshalTOML(t *testing.T) {
	type hftHolder struct {
		HardforkTime *HardforkTime `toml:"time,omitempty"`
	}

	tests := []struct {
		name string
		hft  hftHolder
		want string
	}{
		{
			name: "nil",
			hft:  hftHolder{},
			want: "",
		},
		{
			name: "zero",
			hft:  hftHolder{HardforkTime: new(HardforkTime)},
			want: "time = 0 # Thu 1 Jan 1970 00:00:00 UTC\n",
		},
		{
			name: "non-zero",
			hft:  hftHolder{HardforkTime: NewHardforkTime(1708560000)},
			want: "time = 1708560000 # Thu 22 Feb 2024 00:00:00 UTC\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toml.Marshal(tt.hft)
			require.NoError(t, err)
			require.Equal(t, tt.want, string(got))
		})
	}
}

func TestCopyHardforks(t *testing.T) {
	src := &Hardforks{
		CanyonTime: NewHardforkTime(1),
		DeltaTime:  NewHardforkTime(2),
	}
	dest := &Hardforks{
		CanyonTime: NewHardforkTime(0),
	}

	require.NoError(t, CopyHardforks(src, dest))

	require.Equal(t, &Hardforks{
		CanyonTime: NewHardforkTime(0),
		DeltaTime:  NewHardforkTime(2),
	}, dest)
}
