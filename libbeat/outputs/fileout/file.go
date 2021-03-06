package fileout

import (
	"packetbeat/libbeat/common"
	"packetbeat/libbeat/common/op"
	"packetbeat/libbeat/logp"
	"packetbeat/libbeat/outputs"
)

func init() {
	outputs.RegisterOutputPlugin("file", New)
}

type fileOutput struct {
	beatName string
	rotator  logp.FileRotator
	codec    outputs.Codec
}

// New instantiates a new file output instance.
func New(beatName string, cfg *common.Config, _ int) (outputs.Outputer, error) {
	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	// disable bulk support in publisher pipeline
	cfg.SetInt("flush_interval", -1, -1)
	cfg.SetInt("bulk_max_size", -1, -1)

	output := &fileOutput{beatName: beatName}
	if err := output.init(config); err != nil {
		return nil, err
	}
	return output, nil
}

func (out *fileOutput) init(config config) error {
	var err error

	out.rotator.Path = config.Path
	out.rotator.Name = config.Filename
	if out.rotator.Name == "" {
		out.rotator.Name = out.beatName
	}

	codec, err := outputs.CreateEncoder(config.Codec)
	if err != nil {
		return err
	}

	out.codec = codec

	logp.Info("File output path set to: %v", out.rotator.Path)
	logp.Info("File output base filename set to: %v", out.rotator.Name)

	rotateeverybytes := uint64(config.RotateEveryKb) * 1024
	logp.Info("Rotate every bytes set to: %v", rotateeverybytes)
	out.rotator.RotateEveryBytes = &rotateeverybytes

	keepfiles := config.NumberOfFiles
	logp.Info("Number of files set to: %v", keepfiles)
	out.rotator.KeepFiles = &keepfiles

	err = out.rotator.CreateDirectory()
	if err != nil {
		return err
	}

	err = out.rotator.CheckIfConfigSane()
	if err != nil {
		return err
	}

	return nil
}

// Implement Outputer
func (out *fileOutput) Close() error {
	return nil
}

func (out *fileOutput) PublishEvent(
	sig op.Signaler,
	opts outputs.Options,
	data outputs.Data,
) error {
	var serializedEvent []byte
	var err error

	serializedEvent, err = out.codec.Encode(data.Event)
	if err != nil {
		op.SigCompleted(sig)
		return err
	}

	err = out.rotator.WriteLine(serializedEvent)
	if err != nil {
		if opts.Guaranteed {
			logp.Critical("Unable to write events to file: %s", err)
		} else {
			logp.Err("Error when writing line to file: %s", err)
		}
	}
	op.Sig(sig, err)
	return err
}
