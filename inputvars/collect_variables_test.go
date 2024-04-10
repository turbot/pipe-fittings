package inputvars

import (
	"github.com/turbot/pipe-fittings/app_specific"
	"reflect"
	"testing"
)

func Test_sanitiseVariableNames(t *testing.T) {
	tests := []struct {
		name    string
		src     []byte
		wantSrc []byte
		wantMap map[string]string
	}{
		{
			name:    "unqualified",
			src:     []byte(`var1="foo"`),
			wantSrc: []byte(`var1="foo"`),
			wantMap: map[string]string{},
		},
		{
			name:    "unqualified with spaces",
			src:     []byte(`var1 = "foo"`),
			wantSrc: []byte(`var1 = "foo"`),
			wantMap: map[string]string{},
		},
		{
			name:    "qualified",
			src:     []byte(`m1.var1="foo"`),
			wantSrc: []byte(`____powerpipe_mod_m1_variable_var1____="foo"`),
			wantMap: map[string]string{`____powerpipe_mod_m1_variable_var1____`: "m1.var1"},
		},
		{
			name:    "qualified single spaces",
			src:     []byte(` m1.var1 = "foo"`),
			wantSrc: []byte(` ____powerpipe_mod_m1_variable_var1____ = "foo"`),
			wantMap: map[string]string{`____powerpipe_mod_m1_variable_var1____`: "m1.var1"},
		},
		{
			name:    "qualified multiple spaces",
			src:     []byte(`    m1.var1          = "foo"`),
			wantSrc: []byte(`    ____powerpipe_mod_m1_variable_var1____          = "foo"`),
			wantMap: map[string]string{`____powerpipe_mod_m1_variable_var1____`: "m1.var1"},
		},
	}
	app_specific.AppName = "powerpipe"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, aliasMap := sanitiseVariableNames(tt.src)
			if !reflect.DeepEqual(src, tt.wantSrc) {
				t.Errorf("sanitiseVariableNames() src = %v, wantSrc %v", src, tt.wantSrc)
			}
			if !reflect.DeepEqual(aliasMap, tt.wantMap) {
				t.Errorf("sanitiseVariableNames() aliasMap = %v, wantMap %v", aliasMap, tt.wantMap)
			}
		})
	}
}
