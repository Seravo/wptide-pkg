package payload

import (
	"reflect"
	"testing"

	"github.com/wptide/pkg/message"
	"os"
	"github.com/wptide/pkg/tide"
)

func Test_filePayload_SendPayload(t *testing.T) {

	// Make a /tmp folder
	os.Mkdir("./testdata/tmp", os.ModePerm)

	// Clean up after.
	defer func() {
		os.RemoveAll("./testdata/tmp")
	}()

	terminateChan := make(chan struct{})

	type args struct {
		destination   string
		payload       []byte
		terminateChan chan struct{}
	}
	tests := []struct {
		name    string
		fp      FilePayload
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Write to file",
			FilePayload{},
			args{
				"./testdata/tmp/temp.txt",
				[]byte(`{"title":"","content":"","version":"","checksum":"abcdefg","visibility":"","project_type":"plugin","source_url":"","source_type":"","code_info":{"type":"plugin","details":[],"cloc":{}},"reports":{"phpcs_demo":{"raw":{"type":"mock","filename":"mock","path":"mock"},"parsed":{"type":"mock","filename":"mock","path":"mock"},"summary":{}}}}`),
				nil,
			},
			[]byte("ok"),
			false,
		},
		{
			"Error: Write to directory",
			FilePayload{},
			args{
				"./testdata/tmp",
				[]byte(`Nothing will write`),
				nil,
			},
			nil,
			true,
		},
		{
			"Write to file",
			FilePayload{},
			args{
				"./testdata/tmp/temp.txt",
				[]byte(`{"title":"","content":"","version":"","checksum":"abcdefg","visibility":"","project_type":"plugin","source_url":"","source_type":"","code_info":{"type":"plugin","details":[],"cloc":{}},"reports":{"phpcs_demo":{"raw":{"type":"mock","filename":"mock","path":"mock"},"parsed":{"type":"mock","filename":"mock","path":"mock"},"summary":{}}}}`),
				terminateChan,
			},
			[]byte("ok"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := FilePayload{
				TerminateChannel: tt.args.terminateChan,
			}

			if tt.args.terminateChan != nil {
				go func() {
					for {
						select {
							case <-tt.args.terminateChan:
						}
					}
				}()
			}

			got, err := fp.SendPayload(tt.args.destination, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePayload.SendPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilePayload.SendPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filePayload_BuildPayload(t *testing.T) {
	type args struct {
		msg  message.Message
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		fp      FilePayload
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Get Tide Payload",
			FilePayload{},
			args{
				data: map[string]interface{}{
					"info": tide.CodeInfo{
						"plugin",
						[]tide.InfoDetails{},
						map[string]tide.ClocResult{},
					},
					"phpcs_demo": tide.AuditResult{
						Raw: tide.AuditDetails{
							Type:     "mock",
							FileName: "mock",
							Path:     "mock",
						},
						Parsed: tide.AuditDetails{
							Type:     "mock",
							FileName: "mock",
							Path:     "mock",
						},
					},
					"checksum": "abcdefg",
				},
			},
			[]byte(`{"title":"","content":"","version":"","checksum":"abcdefg","visibility":"","project_type":"plugin","source_url":"","source_type":"","code_info":{"type":"plugin","details":[],"cloc":{}},"reports":{"phpcs_demo":{"raw":{"type":"mock","filename":"mock","path":"mock"},"parsed":{"type":"mock","filename":"mock","path":"mock"},"summary":{}}}}`),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := FilePayload{}
			got, err := fp.BuildPayload(tt.args.msg, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePayload.BuildPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilePayload.BuildPayload() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
