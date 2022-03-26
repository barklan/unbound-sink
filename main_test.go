package main

import (
	"regexp"
	"testing"
)

func Test_parseLine(t *testing.T) {
	type args struct {
		line       string
		badRegexes []*regexp.Regexp
	}
	badRgs, err := badRegexes()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			"empty line",
			args{"", badRgs},
			"",
			false,
		},
		{
			"newline",
			args{"\n", badRgs},
			"",
			false,
		},
		{
			"commented line",
			args{"# This is a comment.", badRgs},
			"",
			false,
		},
		{
			"commented line 2",
			args{"#=====================================", badRgs},
			"",
			false,
		},
		{
			"actual domain",
			args{"0.0.0.0 p.s.360.cn", badRgs},
			"p.s.360.cn",
			true,
		},
		{
			"localhost",
			args{"127.0.0.1 localhost", badRgs},
			"",
			false,
		},
		{
			"localhost 2",
			args{"127.0.0.1 localhost.localdomain", badRgs},
			"",
			false,
		},
		{
			"localhost 3",
			args{"127.0.0.1 local", badRgs},
			"",
			false,
		},
		{
			"weird shit",
			args{"255.255.255.255 broadcasthost", badRgs},
			"",
			false,
		},
		{
			"weird shit 2",
			args{"::1 localhost", badRgs},
			"",
			false,
		},
		{
			"weird shit 3",
			args{"::1 ip6-loopback", badRgs},
			"",
			false,
		},
		{
			"weird shit 3",
			args{"0.0.0.0 0.0.0.0", badRgs},
			"",
			false,
		},
		{
			"weird shit 4",
			args{"fe80::1%lo0 localhost", badRgs},
			"",
			false,
		},
		{
			"another acutal domain",
			args{"r5---sn-n4v7knlz.googlevideo.com", badRgs},
			"r5---sn-n4v7knlz.googlevideo.com",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseLine(tt.args.line, tt.args.badRegexes)
			if got != tt.want {
				t.Errorf("parseLine() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseLine() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_unboundLine(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"simple domain",
			args{"google.com"},
			"local-zone: \"google.com\" always_refuse\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unboundLine(tt.args.domain); got != tt.want {
				t.Errorf("unboundLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseDoc(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"small blocklist",
			args{`
# Title: StevenBlack/hosts
#
# This hosts file is a merged collection of hosts from reputable sources,

127.0.0.1 localhost
127.0.0.1 localhost.localdomain
127.0.0.1 local
255.255.255.255 broadcasthost
::1 localhost
::1 ip6-localhost
::1 ip6-loopback
fe80::1%lo0 localhost
ff00::0 ip6-localnet
ff00::0 ip6-mcastprefix
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
ff02::3 ip6-allhosts
0.0.0.0 0.0.0.0

# Custom host records are listed here.


# End of custom host records.
# Start StevenBlack

#=====================================
# Title: Hosts contributed by Steven Black
# http://stevenblack.com

0.0.0.0 wizhumpgyros.com
0.0.0.0 coccyxwickimp.com
0.0.0.0 webmail-who-int.000webhostapp.com`},
			`local-zone: "wizhumpgyros.com" always_refuse
local-zone: "coccyxwickimp.com" always_refuse
local-zone: "webmail-who-int.000webhostapp.com" always_refuse
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDoc(tt.args.input)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("parseDoc() = %v, want %v", got, tt.want)
			}
		})
	}
}
