package fesgo

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestParseParamsMap(t *testing.T) {

	testCase := []struct {
		name    string
		param   string
		want    map[string]map[string]string
		wantErr error
	}{
		{
			name:    "Parse nil",
			param:   "",
			want:    map[string]map[string]string{},
			wantErr: nil,
		},
		{
			name:  "Parse map",
			param: "user[id]=1&user[name]=张三",
			want: map[string]map[string]string{
				"user": {
					"id":   "1",
					"name": "张三",
				},
			},
			wantErr: nil,
		},
		{
			name:  "Parse multiple layers",
			param: "user[id]=1&user[name]=张三&info[niu]=2",
			want: map[string]map[string]string{
				"user": {
					"id":   "1",
					"name": "张三",
				},
				"info": {
					"niu": "2",
				},
			},
			wantErr: nil,
		},
		{
			name:  "Parse query",
			param: "name=feng&user[id]=1&age=18&user[name]=张三&info[niu]=2",
			want: map[string]map[string]string{
				"user": {
					"id":   "1",
					"name": "张三",
				},
				"info": {
					"niu": "2",
				},
			},
			wantErr: nil,
		},
		{
			name:  "Parse cover",
			param: "user[id]=1&user[name]=张三&info[niu]=2&user[id]=2",
			want: map[string]map[string]string{
				"user": {
					"id":   "2",
					"name": "张三",
				},
				"info": {
					"niu": "2",
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(tt *testing.T) {
			query, err := url.ParseQuery(tc.param)
			if err != nil {
				t.Error("parse err", err)
				return
			}
			paramsMap := ParseParamsMap(query)
			assert.Equal(tt, tc.want, paramsMap)
		})
	}
}
