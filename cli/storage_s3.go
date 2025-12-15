package cli

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/pkg/errors"

	"github.com/kopia/kopia/repo/blob"
	"github.com/kopia/kopia/repo/blob/s3"
)

type storageS3Flags struct {
	s3options       s3.Options
	rootCaPemBase64 string
	rootCaPemPath   string
}

func (c *storageS3Flags) Setup(svc StorageProviderServices, cmd *kingpin.CmdClause) {
	cmd.Flag("bucket", "Name of the S3 bucket").Required().StringVar(&c.s3options.BucketName)
	cmd.Flag("endpoint", "Endpoint to use").Default("s3.amazonaws.com").StringVar(&c.s3options.Endpoint)
	cmd.Flag("region", "S3 Region").Default("").StringVar(&c.s3options.Region)
	cmd.Flag("access-key", "Access key ID (overrides AWS_ACCESS_KEY_ID environment variable)").Required().Envar(svc.EnvName("AWS_ACCESS_KEY_ID")).StringVar(&c.s3options.AccessKeyID)
	cmd.Flag("secret-access-key", "Secret access key (overrides AWS_SECRET_ACCESS_KEY environment variable)").Required().Envar(svc.EnvName("AWS_SECRET_ACCESS_KEY")).StringVar(&c.s3options.SecretAccessKey)
	cmd.Flag("session-token", "Session token (overrides AWS_SESSION_TOKEN environment variable)").Envar(svc.EnvName("AWS_SESSION_TOKEN")).StringVar(&c.s3options.SessionToken)
	cmd.Flag("prefix", "Prefix to use for objects in the bucket. Put trailing slash (/) if you want to use prefix as directory. e.g my-backup-dir/ would put repository contents inside my-backup-dir directory").StringVar(&c.s3options.Prefix)
	cmd.Flag("disable-tls", "Disable TLS security (HTTPS)").BoolVar(&c.s3options.DoNotUseTLS)
	cmd.Flag("disable-tls-verification", "Disable TLS (HTTPS) certificate verification").BoolVar(&c.s3options.DoNotVerifyTLS)

	commonThrottlingFlags(cmd, &c.s3options.Limits)

	var pointInTimeStr string

	pitPreAction := func(_ *kingpin.ParseContext) error {
		if pointInTimeStr != "" {
			t, err := time.Parse(time.RFC3339, pointInTimeStr)
			if err != nil {
				return errors.Wrap(err, "invalid point-in-time argument")
			}

			c.s3options.PointInTime = &t
		}

		return nil
	}

	cmd.Flag("point-in-time", "Use a point-in-time view of the storage repository when supported").PlaceHolder(time.RFC3339).PreAction(pitPreAction).StringVar(&pointInTimeStr)

	cmd.Flag("root-ca-pem-base64", "Certificate authority in-line (base64 enc.)").Envar(svc.EnvName("ROOT_CA_PEM_BASE64")).PreAction(c.preActionLoadPEMBase64).StringVar(&c.rootCaPemBase64)
	cmd.Flag("root-ca-pem-path", "Certificate authority file path").PreAction(c.preActionLoadPEMPath).StringVar(&c.rootCaPemPath)
}

func (c *storageS3Flags) preActionLoadPEMPath(_ *kingpin.ParseContext) error {
	if len(c.s3options.RootCA) > 0 {
		return errors.New("root-ca-pem-base64 and root-ca-pem-path are mutually exclusive")
	}

	data, err := os.ReadFile(c.rootCaPemPath) //#nosec
	if err != nil {
		return errors.Wrapf(err, "error opening root-ca-pem-path %v", c.rootCaPemPath)
	}

	c.s3options.RootCA = data

	return nil
}

func (c *storageS3Flags) preActionLoadPEMBase64(_ *kingpin.ParseContext) error {
	caContent, err := base64.StdEncoding.DecodeString(c.rootCaPemBase64)
	if err != nil {
		return errors.Wrap(err, "unable to decode CA")
	}

	c.s3options.RootCA = caContent

	return nil
}

func (c *storageS3Flags) Connect(ctx context.Context, isCreate bool, formatVersion int) (blob.Storage, error) {
	_ = formatVersion

	if isCreate && c.s3options.PointInTime != nil && !c.s3options.PointInTime.IsZero() {
		return nil, errors.New("Cannot specify a 'point-in-time' option when creating a BSL")
	}

	// OADP: Normalize prefix to include oadp-vmdp/ prefix
	normalizedPrefix, err := normalizeOADPPrefix(c.s3options.Prefix)
	if err != nil {
		return nil, err
	}

	// OADP: Do not mutate c.s3options in-place (prevents double-normalization on repeated calls).
	// Allocate opts explicitly so its lifetime is unambiguous to readers/review tools.
	opts := new(s3.Options)
	*opts = c.s3options
	opts.Prefix = normalizedPrefix

	//nolint:wrapcheck
	return s3.New(ctx, opts, isCreate)
}

// normalizeOADPPrefix prepends "oadp-vmdp/" to the user-provided prefix.
// This ensures OADP data is isolated within shared buckets.
func normalizeOADPPrefix(userPrefix string) (string, error) {
	const oadpPrefix = OADPPrefix // "oadp-vmdp/" from oadp_config.go

	// OADP: Reject leading/trailing whitespace to avoid hard-to-debug prefix mismatches.
	// Internal spaces (e.g. "my backups/") are valid in S3 keys and are allowed.
	if strings.TrimSpace(userPrefix) != userPrefix {
		return "", errors.New("prefix must not start or end with whitespace")
	}

	// OADP: Reject control whitespace which is almost certainly accidental.
	if strings.ContainsAny(userPrefix, "\t\r\n") {
		return "", errors.New("prefix must not contain control whitespace (tabs/newlines)")
	}

	// Clean up any leading slashes from user prefix
	cleanedPrefix := strings.TrimLeft(userPrefix, "/")

	// OADP: Ensure user doesn't include 'oadp-vmdp' as a path segment in their prefix.
	// This prefix segment is automatically added.
	for _, seg := range strings.Split(cleanedPrefix, "/") {
		if seg == "" {
			continue
		}

		if strings.EqualFold(seg, "oadp-vmdp") {
			return "", errors.New("prefix must not contain 'oadp-vmdp' as a path segment - this prefix is automatically added")
		}
	}

	return oadpPrefix + cleanedPrefix, nil
}

func init() {
	mustRegisterStorageProvider(
		"s3",
		"an S3 bucket",
		func() StorageFlags { return &storageS3Flags{} },
	)
}
