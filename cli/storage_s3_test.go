package cli

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kopia/kopia/internal/testutil"
)

func TestNormalizeOADPPrefix(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		want      string
		wantError bool
	}{
		{
			name: "empty-prefix",
			in:   "",
			want: OADPPrefix,
		},
		{
			name: "leading-slash-trimmed",
			in:   "/my-prefix/",
			want: OADPPrefix + "my-prefix/",
		},
		{
			name: "internal-spaces-allowed",
			in:   "my backups/",
			want: OADPPrefix + "my backups/",
		},
		{
			name:      "leading-whitespace-rejected",
			in:        " my-prefix/",
			wantError: true,
		},
		{
			name:      "trailing-whitespace-rejected",
			in:        "my-prefix/ ",
			wantError: true,
		},
		{
			name:      "tab-rejected",
			in:        "my\tprefix/",
			wantError: true,
		},
		{
			name:      "newline-rejected",
			in:        "my-prefix/\n",
			wantError: true,
		},
		{
			name:      "segment-oadp-vmdp-rejected",
			in:        "oadp-vmdp/foo/",
			wantError: true,
		},
		{
			name:      "segment-oadp-vmdp-case-insensitive-rejected",
			in:        "OADP-VMDP/foo/",
			wantError: true,
		},
		{
			name:      "segment-oadp-vmdp-in-middle-rejected",
			in:        "foo/oadp-vmdp/bar/",
			wantError: true,
		},
		{
			name: "substring-not-a-segment-allowed",
			in:   "my-oadp-vmdp-migration/",
			want: OADPPrefix + "my-oadp-vmdp-migration/",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeOADPPrefix(tc.in)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error, got none (result=%q)", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

var (
	fakeCertContent         = []byte("fake certificate content")
	fakeCertContentAsBase64 = base64.StdEncoding.EncodeToString(fakeCertContent)
)

func TestLoadPEMBase64(t *testing.T) {
	var s3flags storageS3Flags

	s3flags = storageS3Flags{rootCaPemBase64: ""}
	require.NoError(t, s3flags.preActionLoadPEMBase64(nil))

	s3flags = storageS3Flags{rootCaPemBase64: "AA=="}
	require.NoError(t, s3flags.preActionLoadPEMBase64(nil))

	s3flags = storageS3Flags{rootCaPemBase64: fakeCertContentAsBase64}
	require.NoError(t, s3flags.preActionLoadPEMBase64(nil))
	require.Equal(t, fakeCertContent, s3flags.s3options.RootCA, "content of RootCA should be %v", fakeCertContent)

	s3flags = storageS3Flags{rootCaPemBase64: "!"}
	err := s3flags.preActionLoadPEMBase64(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "illegal base64 data")
}

func TestLoadPEMPath(t *testing.T) {
	var s3flags storageS3Flags

	tempdir := testutil.TempDirectory(t)
	certpath := filepath.Join(tempdir, "certificate-filename")

	require.NoError(t, os.WriteFile(certpath, fakeCertContent, 0o644))

	// Test regular file
	s3flags = storageS3Flags{rootCaPemPath: certpath}
	require.NoError(t, s3flags.preActionLoadPEMPath(nil))
	require.Equal(t, fakeCertContent, s3flags.s3options.RootCA, "content of RootCA should be %v", fakeCertContent)

	// Test inexistent file
	s3flags = storageS3Flags{rootCaPemPath: "/does-not-exists"}
	err := s3flags.preActionLoadPEMPath(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error opening root-ca-pem-path")
}

func TestLoadPEMBoth(t *testing.T) {
	s3flags := storageS3Flags{rootCaPemBase64: "AA==", rootCaPemPath: "/tmp/blah"}
	require.NoError(t, s3flags.preActionLoadPEMBase64(nil))
	err := s3flags.preActionLoadPEMPath(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "mutually exclusive")
}
