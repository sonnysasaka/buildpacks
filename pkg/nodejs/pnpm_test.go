// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package nodejs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/buildpacks/internal/testserver"
	"github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
	"github.com/buildpacks/libcnb"
)

func TestInstallPNPM(t *testing.T) {
	testCases := []struct {
		name        string
		npmResponse string
		packageJSON PackageJSON
		wantFile    string
		wantError   bool
	}{
		{
			name:     "no version constraint",
			wantFile: "bin/pnpm",
			npmResponse: `{
				"name": "pnpm",
				"dist-tags": {
					"latest": "8.4.0"
				},
				"versions": {
					"8.4.0": {
						"name": "npm",
						"version": "8.4.0"
					}
				},
				"modified": "2022-01-27T21:10:55.626Z"
			}`,
			packageJSON: PackageJSON{},
		},
		{
			name:     "valid version constraint",
			wantFile: "bin/pnpm",
			npmResponse: `{
				"name": "pnpm",
				"dist-tags": {
					"latest": "8.4.0"
				},
				"versions": {
					"8.4.0": {
						"name": "npm",
						"version": "8.4.0"
					}
				},
				"modified": "2022-01-27T21:10:55.626Z"
			}`,
			packageJSON: PackageJSON{
				Engines: packageEnginesJSON{
					PNPM: "8.x.x",
				},
			},
		},
		{
			name: "invalid version",
			npmResponse: `{
				"name": "pnpm",
				"dist-tags": {
					"latest": "8.4.0"
				},
				"versions": {
					"8.4.0": {
						"name": "npm",
						"version": "8.4.0"
					}
				},
				"modified": "2022-01-27T21:10:55.626Z"
			}`,
			packageJSON: PackageJSON{
				Engines: packageEnginesJSON{
					PNPM: ">9.0.0",
				},
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testserver.New(
				t,
				testserver.WithJSON(`pnpm!`),
				testserver.WithMockURL(&pnpmDownloadURL),
			)
			testserver.New(
				t,
				testserver.WithJSON(tc.npmResponse),
				testserver.WithMockURL(&npmRegistryURL),
			)

			layer := &libcnb.Layer{
				Name:     "pnpm_test",
				Path:     t.TempDir(),
				Metadata: map[string]any{},
			}
			err := InstallPNPM(gcpbuildpack.NewContext(), layer, &tc.packageJSON)
			if tc.wantError == (err == nil) {
				t.Fatalf("InstallPNPM() got error: %v, want error? %v", err, tc.wantError)
			}

			if tc.wantFile != "" {
				fp := filepath.Join(layer.Path, tc.wantFile)
				if _, err := os.Stat(fp); err != nil {
					t.Errorf("Missing file: %s (%v)", fp, err)
				}
			}
		})
	}
}
