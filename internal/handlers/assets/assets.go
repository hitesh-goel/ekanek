package assets

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hitesh-goel/ekanek/internal/handlers/auth"
	"github.com/hitesh-goel/ekanek/internal/handlers/response"
	awss3 "github.com/hitesh-goel/ekanek/internal/pkg/aws"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type AssetResources struct {
	Session *session.Session
	DTO     *sql.DB
}

type CreateAsset struct {
	Title       string `db:"title"`
	Description string `db:"description"`
	Name        string `db:"name"`
	Public      bool   `db:"public"`
	UserId      string `db:"uid"`
	Path        string `db:"s3_path"`
}

type Asset struct {
	Id          string `json:"asset_id" db:"id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	Name        string `json:"asset_name" db:"name"`
	Public      bool   `json:"is_public" db:"public"`
	UserId      string `json:"uid" db:"uid"`
	Path        string `json:"s3_path" db:"s3_path"`
}

func HandleAssetUpload(ar *AssetResources) (string, func(http.ResponseWriter, *http.Request)) {
	return "/api/v1/asset/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		uploadFile(w, r, ar)
	}
}

func HandleAssetDownload(ar *AssetResources) (string, func(http.ResponseWriter, *http.Request)) {
	return "/api/v1/asset/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		downloadFile(w, r, ar)
	}
}

func HandleListAssets(ar *AssetResources) (string, func(http.ResponseWriter, *http.Request)) {
	return "/api/v1/asset/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		getUserAssets(w, r, ar)
	}
}

func HandlePublicAsset(ar *AssetResources) (string, func(http.ResponseWriter, *http.Request)) {
	return "/api/v1/asset/public", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		grantPublicAccess(w, r, ar)
	}
}

func HandleDeleteAsset(ar *AssetResources) (string, func(http.ResponseWriter, *http.Request)) {
	return "/api/v1/asset/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			response.RespondWithError(w, r, "wrong http method", http.StatusMethodNotAllowed)
			return
		}
		deleteAsset(w, r, ar)
	}
}

func downloadFile(w http.ResponseWriter, r *http.Request, ar *AssetResources) {
	queryValues := r.URL.Query()
	assetId := queryValues.Get("asset_id")

	if assetId == "" {
		response.RespondWithError(w, r, "pass valid asset_id in query param", http.StatusBadRequest)
		return
	}

	var asset CreateAsset
	var query = `select public, s3_path, name from assets where id = $1`
	row := ar.DTO.QueryRow(query, assetId)
	err := row.Scan(&asset.Public, &asset.Path, &asset.Name)
	if err != nil {
		log.Println("Database error", err.Error())
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	if !asset.Public {
		userId, err := auth.Verify(r)
		if err != nil || !strings.HasPrefix(asset.Path, userId) {
			log.Println("user not authorised")
			response.RespondWithError(w, r, "you are not authenticated to access this asset", http.StatusForbidden)
			return
		}
	}

	f, err := os.Create(assetId)
	if err != nil {
		log.Println("Local File Creation Error", err.Error())
		response.RespondWithError(w, r, "Something went wrong creating the local file", http.StatusBadRequest)
		return
	}

	err = awss3.DownloadFromS3(asset.Path, f, ar.Session)
	if err != nil {
		log.Println("S3 download Error: ", err.Error())
		response.RespondWithError(w, r, "Something went wrong retrieving the file from S3", http.StatusBadRequest)
		return
	}

	reader := unCompressFile(f)

	defer f.Close()
	defer os.Remove(assetId)

	w.Header().Set("Content-Disposition", "attachment; filename="+asset.Name)
	_, _ = io.Copy(w, reader)
}

func uploadFile(w http.ResponseWriter, r *http.Request, ar *AssetResources) {
	ctx := r.Context()
	http.MaxBytesReader(w, r.Body, 1024<<20) // request body should not be greater than 1GB

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Error while reading FormFile: ", err.Error())
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	//Detect Content Type of the file uploaded
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		log.Println("Error while reading file buffer", err.Error())
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(buffer)

	if !strings.HasPrefix(contentType, "image") {
		response.RespondWithError(w, r, "Not a valid Video Format File", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	defer file.Close()

	uid, err := auth.GetUID(ctx)
	if err != nil {
		log.Println("Error accessing userId", err.Error())
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := handler.Filename
	compressed := compressFile(file)
	s3Key := fmt.Sprintf("%s/%s.gz", uid, fileName)

	asset := CreateAsset{
		Title:       title,
		Description: description,
		Name:        fileName,
		UserId:      uid,
		Path:        s3Key,
	}

	//TODO: Implement transactions support so that we should insert a new record only after a successful upload
	var query = `
		WITH uuid AS (
			SELECT * FROM uuid_generate_v1mc()
		)
		INSERT INTO assets (
			id,
            uid,
            name,
            s3_path,
            title,
            description
        ) VALUES (
			(SELECT * FROM uuid),
            $1,
            $2,
            $3,
            $4,
            $5
        ) RETURNING id`
	fileId := ""
	err = ar.DTO.QueryRow(query, asset.UserId, asset.Name, asset.Path, asset.Title, asset.Description).Scan(&fileId)
	if err != nil {
		log.Println("Error Inserting record to postgres: ", err.Error())
		response.RespondWithError(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}

	//TODO: should error out on same filename or title?
	_, err = awss3.SaveToS3(s3Key, compressed, ar.Session)
	if err != nil {
		log.Println("Error while uploading file to s3", err.Error())
		response.RespondWithError(w, r, "failed to upload", http.StatusInternalServerError)
		return
	}

	response.RespondWithSuccess(w, r, "Successfully uploaded", "", http.StatusOK)
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

func unCompressFile(f *os.File) *io.PipeReader {
	reader, writer := io.Pipe()
	go func() {
		gw, _ := gzip.NewReader(f)
		_, _ = io.Copy(writer, gw)
		defer writer.Close()
	}()
	return reader
}

func getUserAssets(w http.ResponseWriter, r *http.Request, ar *AssetResources) {
	ctx := r.Context()
	uid, err := auth.GetUID(ctx)
	if err != nil {
		log.Println("Error accessing UserID: ", err.Error())
		response.RespondWithError(w, r, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var query = `select id, uid, title, description, name, s3_path, public from assets where uid = $1`
	rows, err := ar.DTO.Query(query, uid)
	if err != nil {
		log.Println("Error selecting postgres record", err.Error())
		response.RespondWithError(w, r, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var data []Asset
	for rows.Next() {
		var res Asset
		err = rows.Scan(&res.Id, &res.UserId, &res.Title, &res.Description, &res.Name, &res.Path, &res.Public)
		if err != nil {
			log.Println("Error while scanning Asset Rows: ", err.Error())
			response.RespondWithError(w, r, "database error", http.StatusInternalServerError)
			return
		}
		data = append(data, res)
	}

	response.RespondWithSuccess(w, r, "List", data, http.StatusOK)
}

func grantPublicAccess(w http.ResponseWriter, r *http.Request, ar *AssetResources) {
	var asset Asset
	err := json.NewDecoder(r.Body).Decode(&asset)
	if err != nil {
		log.Println("Error decoding json request: ", err.Error())
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	var query = `UPDATE assets set public = true where id = $1`
	rows, err := ar.DTO.Query(query, asset.Id)
	defer rows.Close()
	if err != nil {
		log.Println("Error executing query", err.Error())
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	//TODO: build a tiny URL and save it to DB

	response.RespondWithSuccess(w, r, "Successfully granted public access", "", http.StatusOK)
}

func deleteAsset(w http.ResponseWriter, r *http.Request, ar *AssetResources) {
	var asset Asset
	err := json.NewDecoder(r.Body).Decode(&asset)
	if err != nil {
		log.Println("Error decoding json data", err.Error())
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	var query = `UPDATE assets set is_active = false where id = $1`
	rows, err := ar.DTO.Query(query, asset.Id)
	defer rows.Close()
	if err != nil {
		log.Println("Error updating asset record", err.Error())
		response.RespondWithError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO: build a worker which will access the deleted files from is_active flag and deletes from the s3 later after 30 days or a set duration.
	response.RespondWithSuccess(w, r, "success", "", http.StatusOK)
}

// TODO: remove public access for the asset

// TODO: Grant public access to an asset
//| asset_id | tiny_url | s3_path | is_active | created_at | updated_at
// 1. update public access in assets table
// 2. generate a tiny url and save in tiny_urls table
// 3. return the tiny url path
// 4. download of asset can happen in 2 forms
// 4.1 through tiny url
// 4.1.1 if tiny url not found return 404
// 4.1.2 if tiny url found validate if it is active or not
// 4.1.3 if tiny url is valid validate it witch assets table if the asset is public or not
// 4.1.4 if the asset is public download the file
// 4.2 logged in user ask for asset download
// 4.2.1 API request with JWT to get the asset
// 4.2.2 GET the asset path from s3 and download it.
