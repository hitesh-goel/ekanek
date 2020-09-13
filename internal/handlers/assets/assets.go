package assets

import (
	"compress/gzip"
	"database/sql"
	"github.com/hitesh-goel/ekanek/internal/handlers/response"
	awss3 "github.com/hitesh-goel/ekanek/internal/pkg/aws"
	"io"
	"net/http"
)

func HandleAssetUpload(db *sql.DB, sess awss3.AwsResources) (string, func(http.ResponseWriter, *http.Request)) {
	return "/asset/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		UploadFile(w, r, sess)
	}
}

func UploadFile(w http.ResponseWriter, r *http.Request, sess awss3.AwsResources) {
	http.MaxBytesReader(w, r.Body, 1024<<20) // request body should bot be greater than 1GB

	file, handler, err := r.FormFile("file")
	if err != nil {
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	compressed := compressFile(file)

	location, err := sess.SaveToS3("user/"+handler.Filename+".zip", compressed)

	if err != nil {
		response.RespondWithError(w, r, "failed to upload "+err.Error(), http.StatusInternalServerError)
		return
	}
	response.RespondWithStatus(w, r, "Successfully uploaded to: "+location, http.StatusOK)
}

func compressFile(srcFile io.Reader) *io.PipeReader {
	reader, writer := io.Pipe()
	go func() {
		gw := gzip.NewWriter(writer)
		_, _ = io.Copy(gw, srcFile)
		_ = gw.Close()
		_ = writer.Close()
	}()
	return reader
}
