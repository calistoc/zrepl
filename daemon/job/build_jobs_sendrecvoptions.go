package job

import (
	"github.com/pkg/errors"

	"github.com/zrepl/zrepl/config"
	"github.com/zrepl/zrepl/daemon/filters"
	"github.com/zrepl/zrepl/endpoint"
	"github.com/zrepl/zrepl/util/nodefault"
	"github.com/zrepl/zrepl/zfs"
)

type SendingJobConfig interface {
	GetFilesystems() config.FilesystemsFilter
	GetSendOptions() *config.SendOptions // must not be nil
	GetJobSnapPrefix() string
}

func buildSenderConfig(in SendingJobConfig, jobID endpoint.JobID) (*endpoint.SenderConfig, error) {

	fsf, err := filters.DatasetMapFilterFromConfig(in.GetFilesystems())
	if err != nil {
		return nil, errors.Wrap(err, "cannot build filesystem filter")
	}
	sendOpts := in.GetSendOptions()
	return &endpoint.SenderConfig{
		FSF:           fsf,
		JobID:         jobID,
		JobSnapPrefix: in.GetJobSnapPrefix(),

		Encrypt:              &nodefault.Bool{B: sendOpts.Encrypted},
		SendRaw:              sendOpts.Raw,
		SendProperties:       sendOpts.SendProperties,
		SendBackupProperties: sendOpts.BackupProperties,
		SendLargeBlocks:      sendOpts.LargeBlocks,
		SendCompressed:       sendOpts.Compressed,
		SendEmbeddedData:     sendOpts.EmbeddedData,
		SendSaved:            sendOpts.Saved,
	}, nil
}

type ReceivingJobConfig interface {
	GetRootFS() string
	GetAppendClientIdentity() bool
	GetRecvOptions() *config.RecvOptions
}

func buildReceiverConfig(in ReceivingJobConfig, jobID endpoint.JobID) (rc endpoint.ReceiverConfig, err error) {
	rootFs, err := zfs.NewDatasetPath(in.GetRootFS())
	if err != nil {
		return rc, errors.New("root_fs is not a valid zfs filesystem path")
	}
	if rootFs.Length() <= 0 {
		return rc, errors.New("root_fs must not be empty") // duplicates error check of receiver
	}

	recvOpts := in.GetRecvOptions()
	rc = endpoint.ReceiverConfig{
		JobID:                      jobID,
		RootWithoutClientComponent: rootFs,
		AppendClientIdentity:       in.GetAppendClientIdentity(),

		InheritProperties:  recvOpts.Properties.Inherit,
		OverrideProperties: recvOpts.Properties.Override,
	}
	if err := rc.Validate(); err != nil {
		return rc, errors.Wrap(err, "cannot build receiver config")
	}

	return rc, nil
}
