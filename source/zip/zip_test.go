package zip

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"errors"
)

var fileServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	switch r.URL.String() {
	case "/test.zip":
		w.Header().Set("Content-Type", "applicaiton/zip")
		w.Header().Set("Content-Disposition", "attachment; filename='test.zip'")
		http.ServeFile(w, r, "./testdata/test.zip")
	}
}))

func Test_combinedChecksum(t *testing.T) {
	type args struct {
		sums []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Successful Sum (Sorted)",
			args{
				[]string{
					"27dd8ed44a83ff94d557f9fd0412ed5a8cbca69ea04922d88c01184a07300a5a",
					"2c8b08da5ce60398e1f19af0e5dccc744df274b826abe585eaba68c525434806",
					"f6936912184481f5edd4c304ce27c5a1a827804fc7f329f43d273b8621870776",
				},
			},
			"5a0c0a95d189c266ca1ed43767dd98f3fb513ce3434e2b08f34828ac11e79a94",
		},
		{
			"Successful Sum (Unsorted)",
			args{
				[]string{
					"f6936912184481f5edd4c304ce27c5a1a827804fc7f329f43d273b8621870776",
					"27dd8ed44a83ff94d557f9fd0412ed5a8cbca69ea04922d88c01184a07300a5a",
					"2c8b08da5ce60398e1f19af0e5dccc744df274b826abe585eaba68c525434806",
				},
			},
			"5a0c0a95d189c266ca1ed43767dd98f3fb513ce3434e2b08f34828ac11e79a94",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := combinedChecksum(tt.args.sums); got != tt.want {
				t.Errorf("combinedChecksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZip_GetChecksum(t *testing.T) {

	checksum := "5a0c0a95d189c266ca1ed43767dd98f3fb513ce3434e2b08f34828ac11e79a94"

	// Should be impossible to fail.
	t.Run("Get Checksum", func(t *testing.T) {
		m := Zip{
			checksum: checksum,
		}
		if got := m.GetChecksum(); got != checksum {
			t.Errorf("Zip.GetChecksum() = %v, want %v", got, checksum)
		}
	})
}

func TestZip_GetFiles(t *testing.T) {

	files := []string{
		"file1.txt",
		"file2.txt",
		"file3.txt",
	}

	// Should be impossible to fail.
	t.Run("Get Files", func(t *testing.T) {
		m := Zip{
			files: files,
		}
		if got := m.GetFiles(); !reflect.DeepEqual(got, files) {
			t.Errorf("Zip.GetFiles() = %v, want %v", got, files)
		}
	})
}

func Test_unzip(t *testing.T) {

	// Clean up after.
	defer func() {
		os.RemoveAll("./testdata/unzipped")
	}()

	errorDirectoryCreate := func(path string, perm os.FileMode) error {
		return errors.New("something went wrong.")
	}

	errorCopySHA := func(dst io.Writer, src io.Reader) (written int64, err error) {
		if reflect.TypeOf(dst).String() != "*sha256.digest" {
			return io.Copy(dst, src)
		}
		return 0, errors.New("something went wrong")
	}

	errorCopyFile := func(dst io.Writer, src io.Reader) (written int64, err error) {
		if reflect.TypeOf(dst).String() != "*os.File" {
			return io.Copy(dst, src)
		}
		return 0, errors.New("something went wrong")
	}

	errorOpenFile := func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return nil, errors.New("something went wrong")
	}

	type args struct {
		source           string
		destination      string
		makeDirectoryAll func(path string, perm os.FileMode) error
		ioCopy           func(dst io.Writer, src io.Reader) (written int64, err error)
		openFile         func(name string, flag int, perm os.FileMode) (*os.File, error)
	}
	tests := []struct {
		name          string
		args          args
		wantFilenames []string
		wantChecksums []string
		wantErr       bool
	}{
		{
			"Unzip File - Success",
			args{
				source:      "./testdata/test.zip",
				destination: "./testdata/unzipped",
			},
			[]string{
				"testdata/unzipped/function.php",
				"testdata/unzipped/script.js",
				"testdata/unzipped/style.css",
			},
			[]string{
				"64a43b6ce686b50bbd7eb91b2b1346ed66e7053d42f7f7b9d5562d55a25d1321",
				"9a8549c5d1f384593788dc25b1c236f8450534e8cb95833003786fef8201b92b",
				"09679b8abb88b21dd1cf166e1d2745df7882a879d2b8672548f6dc0dc9572fe6",
			},
			false,
		},
		{
			"Unzip File - File",
			args{
				source:      "./testdata/error.zip",
				destination: "./testdata/unzipped",
			},
			nil,
			nil,
			true,
		},
		{
			"Unzip - Failed Directory Create",
			args{
				source:           "./testdata/test.zip",
				destination:      "./testdata/unzipped",
				makeDirectoryAll: errorDirectoryCreate,
			},
			nil,
			nil,
			true,
		},
		{
			"Unzip - Faile Copy to Hasher",
			args{
				source:      "./testdata/test.zip",
				destination: "./testdata/unzipped",
				ioCopy:      errorCopySHA,
			},
			nil,
			nil,
			true,
		},
		{
			"Unzip - Fail Open Target File",
			args{
				source:      "./testdata/test.zip",
				destination: "./testdata/unzipped",
				openFile:    errorOpenFile,
			},
			nil,
			nil,
			true,
		},
		{
			"Unzip - Faile Copy to Target File",
			args{
				source:      "./testdata/test.zip",
				destination: "./testdata/unzipped",
				ioCopy:      errorCopyFile,
			},
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Test create file.
			if tt.args.makeDirectoryAll != nil {
				oldMakeDirectoryAll := makeDirectoryAll
				makeDirectoryAll = tt.args.makeDirectoryAll
				defer func() {
					makeDirectoryAll = oldMakeDirectoryAll
				}()
			}

			// Test io.Copy error.
			if tt.args.ioCopy != nil {
				oldCopy := ioCopy
				ioCopy = tt.args.ioCopy
				defer func() {
					ioCopy = oldCopy
				}()
			}

			// Test io.Copy error.
			if tt.args.openFile != nil {
				oldOpenFile := openFile
				openFile = tt.args.openFile
				defer func() {
					openFile = oldOpenFile
				}()
			}

			gotFilenames, gotChecksums, err := unzip(tt.args.source, tt.args.destination)
			if (err != nil) != tt.wantErr {
				t.Errorf("unzip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFilenames, tt.wantFilenames) {
				t.Errorf("unzip() gotFilenames = %v, want %v", gotFilenames, tt.wantFilenames)
			}
			if !reflect.DeepEqual(gotChecksums, tt.wantChecksums) {
				t.Errorf("unzip() gotChecksums = %v, want %v", gotChecksums, tt.wantChecksums)
			}
		})
	}
}

func TestZip_PrepareFiles(t *testing.T) {

	dest := "./testdata/download/"
	errDest := "./testdata/error/"

	// Clean up after.
	defer func() {
		os.RemoveAll(dest)
		os.RemoveAll(errDest)
	}()

	errorCreate := func(path string) (*os.File, error) {
		return nil, errors.New("something went wrong")
	}

	type fields struct {
		url      string
		dest     string
		files    []string
		checksum string
	}
	type args struct {
		dest           string
		createFile     func(string) (*os.File, error)
		sourceFilename string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Simple Zip Fetch",
			fields{
				url:  fileServer.URL + "/test.zip",
				dest: dest,
			},
			args{
				dest: dest,
			},
			false,
		},
		{
			"Error Destination",
			fields{
				url:  fileServer.URL + "/test.zip",
				dest: dest,
			},
			args{
				dest:       dest,
				createFile: errorCreate,
			},
			true,
		},
		{
			"Error Source Zip",
			fields{
				url:  fileServer.URL + "/error.zip",
				dest: errDest,
			},
			args{
				dest:           errDest,
				sourceFilename: "error.zip",
			},
			true,
		},
		{
			"Error Url",
			fields{
				url:  "https://error.err/error.zip",
				dest: dest,
			},
			args{
				dest: dest,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Test create file.
			if tt.args.createFile != nil {
				oldCreateFile := createFile
				createFile = tt.args.createFile
				defer func() {
					createFile = oldCreateFile
				}()
			}

			// Test bad source name.
			if tt.args.sourceFilename != "" {
				oldFilename := sourceFilename
				sourceFilename = tt.args.sourceFilename
				defer func() {
					sourceFilename = oldFilename
				}()
			}

			m := &Zip{
				url:      tt.fields.url,
				dest:     tt.fields.dest,
				files:    tt.fields.files,
				checksum: tt.fields.checksum,
			}
			if err := m.PrepareFiles(tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("Zip.PrepareFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_downloadFile(t *testing.T) {

	dest := "./testdata/source.zip"

	// Clean up after.
	defer func() {
		os.Remove(dest)
	}()

	errorCopy := func(dst io.Writer, src io.Reader) (written int64, err error) {
		return 0, errors.New("something went wrong")
	}

	type args struct {
		source      string
		destination string
		ioCopy      func(dst io.Writer, src io.Reader) (written int64, err error)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Download - Success",
			args{
				source:      fileServer.URL + "/test.zip",
				destination: dest,
			},
			false,
		},
		{
			"Download - Fail Copy to Target",
			args{
				source:      fileServer.URL + "/test.zip",
				destination: dest,
				ioCopy:      errorCopy,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Test io.Copy error.
			if tt.args.ioCopy != nil {
				oldCopy := ioCopy
				ioCopy = tt.args.ioCopy
				defer func() {
					ioCopy = oldCopy
				}()
			}

			if err := downloadFile(tt.args.source, tt.args.destination); (err != nil) != tt.wantErr {
				t.Errorf("downloadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewZip(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want *Zip
	}{
		{
			"Get new *Zip",
			args{
				fileServer.URL + "/test.zip",
			},
			&Zip{
				url: fileServer.URL + "/test.zip",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewZip(tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewZip() = %v, want %v", got, tt.want)
			}
		})
	}
}
