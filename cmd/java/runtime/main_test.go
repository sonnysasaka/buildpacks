// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"reflect"
	"testing"

	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
)

func TestDetect(t *testing.T) {
	testCases := []struct {
		name  string
		files map[string]string
		want  int
	}{
		{
			name: "java files",
			files: map[string]string{
				"main.java": "",
			},
			want: 0,
		},
		{
			name:  "no java files",
			files: map[string]string{},
			want:  100,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gcp.TestDetect(t, detectFn, tc.name, tc.files, []string{}, tc.want)
		})
	}
}

func TestParseVersionJSON(t *testing.T) {
	testCases := []struct {
		name         string
		json         string
		wantVersion  string
		wantBinaries []binary
	}{
		{
			name: "1 release",
			json: `[{
  "release_name": "jdk-11.0.6+10",
  "binaries": [
    {
      "os": "linux",
      "architecture": "x64",
      "binary_type": "jdk",
      "binary_link": "https://example.com/want"
    }
  ]
}]`,
			wantVersion: "jdk-11.0.6+10",
			wantBinaries: []binary{
				binary{
					BinaryLink:   "https://example.com/want",
					BinaryType:   "jdk",
					OS:           "linux",
					Architecture: "x64",
				},
			},
		},
		{
			name: "2 releases",
			json: `[{
  "release_name": "jdk-11.0.5+10",
  "binaries": [
    {
      "os": "linux",
      "architecture": "x64",
      "binary_type": "jdk",
      "binary_link": "https://example.com/want"
    }
  ]
},
{
	"release_name": "jdk-11.0.6+10",
	"binaries": [
		{
			"os": "linux",
			"architecture": "x64",
			"binary_type": "jdk",
			"binary_link": "https://example2.com/want"
		}
	]
}]`,
			wantVersion: "jdk-11.0.6+10",
			wantBinaries: []binary{
				binary{
					BinaryLink:   "https://example2.com/want",
					BinaryType:   "jdk",
					OS:           "linux",
					Architecture: "x64",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			release, err := parseVersionJSON(tc.json)
			if err != nil {
				t.Fatalf("parseUserVersionJSON() returned error: %v", err)
			}
			if release.Version != tc.wantVersion {
				t.Errorf("release version from parseVersionJSON()=%s, want=%s", release.Version, tc.wantVersion)
			}
			if !reflect.DeepEqual(release.Binaries, tc.wantBinaries) {
				t.Errorf("binaries from parseVersionJSON()=%v, want=%v", release.Binaries, tc.wantBinaries)
			}
		})
	}
}

func TestParseVersionJSONFail(t *testing.T) {
	testCases := []struct {
		name string
		json string
	}{
		{
			name: "invalid JSON",
			json: `[{]`,
		},
		{
			name: "0 releases",
			json: `[]`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseVersionJSON(tc.json)
			if err == nil {
				t.Error("parseVersionJSON() did not return error.")
			}
		})
	}
}

func TestExtractRelease(t *testing.T) {
	testCases := []struct {
		name           string
		javaRelease    javaRelease
		wantVersion    string
		wantBinaryLink string
	}{
		{
			name: "1 binary",
			javaRelease: javaRelease{
				Version: "jdk-11.0.6+10",
				Binaries: []binary{
					binary{
						BinaryLink:   "https://example.com/want",
						BinaryType:   "jdk",
						OS:           "linux",
						Architecture: "x64",
					},
				},
			},
			wantVersion:    "11.0.6+10",
			wantBinaryLink: "https://example.com/want",
		},
		{
			name: "2 binaries with wrong binary type",
			javaRelease: javaRelease{
				Version: "jdk-11.0.6+10",
				Binaries: []binary{
					binary{
						BinaryLink:   "https://example.com/want",
						BinaryType:   "jre",
						OS:           "linux",
						Architecture: "x64",
					},
					binary{
						BinaryLink:   "https://example2.com/want",
						BinaryType:   "jdk",
						OS:           "linux",
						Architecture: "x64",
					},
				},
			},
			wantVersion:    "11.0.6+10",
			wantBinaryLink: "https://example2.com/want",
		},
		{
			name: "2 binaries with wrong OS",
			javaRelease: javaRelease{
				Version: "jdk-11.0.6+10",
				Binaries: []binary{
					binary{
						BinaryLink:   "https://example.com/want",
						BinaryType:   "jdk",
						OS:           "windows",
						Architecture: "x64",
					},
					binary{
						BinaryLink:   "https://example2.com/want",
						BinaryType:   "jdk",
						OS:           "linux",
						Architecture: "x64",
					},
				},
			},
			wantVersion:    "11.0.6+10",
			wantBinaryLink: "https://example2.com/want",
		},
		{
			name: "2 binaries with wrong architecture",
			javaRelease: javaRelease{
				Version: "jdk-11.0.6+10",
				Binaries: []binary{
					binary{
						BinaryLink:   "https://example.com/want",
						BinaryType:   "jdk",
						OS:           "linux",
						Architecture: "x86",
					},
					binary{
						BinaryLink:   "https://example2.com/want",
						BinaryType:   "jdk",
						OS:           "linux",
						Architecture: "x64",
					},
				},
			},
			wantVersion:    "11.0.6+10",
			wantBinaryLink: "https://example2.com/want",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotVersion, gotBinaryLink, err := extractRelease(tc.javaRelease)
			if err != nil {
				t.Fatalf("extractRelease() returned error: %v", err)
			}
			if gotVersion != tc.wantVersion {
				t.Errorf("release version from extractRelease()=%s, want=%s", gotVersion, tc.wantVersion)
			}
			if gotBinaryLink != tc.wantBinaryLink {
				t.Errorf("binaries from extractRelease()=%v, want=%v", gotBinaryLink, tc.wantBinaryLink)
			}
		})
	}
}

func TestExtractReleaseFail(t *testing.T) {
	testCases := []struct {
		name        string
		javaRelease javaRelease
	}{
		{
			name: "0 binaries",
			javaRelease: javaRelease{
				Version:  "jdk-11.0.6+10",
				Binaries: []binary{},
			},
		},
		{
			name: "binaries with wrong binary fields",
			javaRelease: javaRelease{
				Version: "jdk-11.0.6+10",
				Binaries: []binary{
					binary{
						BinaryLink:   "https://example.com/want",
						BinaryType:   "jre",
						OS:           "linux",
						Architecture: "x64",
					},
					binary{
						BinaryLink:   "https://example2.com/want",
						BinaryType:   "jdk",
						OS:           "windows",
						Architecture: "x64",
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := extractRelease(tc.javaRelease)
			if err == nil {
				t.Error("extractRelease() did not return error.")
			}
		})
	}
}