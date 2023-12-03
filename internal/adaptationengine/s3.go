package adaptationengine

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	sascomv1 "github.com/yashvardhan-kukreja/consulkv-commander/api/v1"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/utils"
	"net/http"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

func parseBucketName(link string) (string, string, bool) {
	//https://stryds-media.s3.amazonaws.com/secret_invalidations_tracker.csv
	if link == "" {
		return "", "", false
	}
	parsedURL, err := url.Parse(link)
	if err != nil {
		return "", "", false
	}

	hostSegments := strings.Split(parsedURL.Host, ".")
	if len(hostSegments) < 2 {
		return "", "", false
	}
	bucketName := hostSegments[0]
	objectKey := strings.TrimLeft(parsedURL.Path, "/")

	return bucketName, objectKey, true
}

func (s Client) adaptSheet(item *sascomv1.KVGroup, newInvalidationsOutput utils.InvalidationsOutput) error {
	bucketName, objectKey, ok := parseBucketName(s.sheetLink)
	if !ok {
		return nil
	}

	targetKvGroupKey := client.ObjectKeyFromObject(item).String()

	resp, statusCode, err := utils.CallAPI(utils.APIRequest{
		URL:    s.sheetLink,
		Method: utils.GET,
	})
	if err != nil {
		return fmt.Errorf("error occurred while downloading the sheet associated with the provided link: %w", err)
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("unable to access the provided link(status code: %d): %s", statusCode, string(resp))
	}

	reader := csv.NewReader(bytes.NewReader(resp))
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error occurred while reading the records of the CSV: %w", err)
	}

	kvGroupHeaderIdx, timeHeaderIdx, detailsHeaderIdx, linkHeaderIdx := -1, -1, -1, -1
	if len(records) == 0 {
		kvGroupHeaderIdx = 0
		records = [][]string{{"Link", "Time", "Details", "Link"}}
	} else {
		headerRecord := records[0]
		for idx, col := range headerRecord {
			switch col {
			case "KVGroup":
				kvGroupHeaderIdx = idx
			case "Time":
				timeHeaderIdx = idx
			case "Details":
				detailsHeaderIdx = idx
			case "Link":
				linkHeaderIdx = idx
			}
		}
	}
	if kvGroupHeaderIdx == -1 {
		curMax, _ := utils.MaxInSlice([]int{kvGroupHeaderIdx, timeHeaderIdx, detailsHeaderIdx, linkHeaderIdx})
		kvGroupHeaderIdx = 1 + curMax
	}
	if timeHeaderIdx == -1 {
		curMax, _ := utils.MaxInSlice([]int{kvGroupHeaderIdx, timeHeaderIdx, detailsHeaderIdx, linkHeaderIdx})
		timeHeaderIdx = 1 + curMax
	}
	if detailsHeaderIdx == -1 {
		curMax, _ := utils.MaxInSlice([]int{kvGroupHeaderIdx, timeHeaderIdx, detailsHeaderIdx, linkHeaderIdx})
		detailsHeaderIdx = 1 + curMax
	}
	if linkHeaderIdx == -1 {
		curMax, _ := utils.MaxInSlice([]int{kvGroupHeaderIdx, timeHeaderIdx, detailsHeaderIdx, linkHeaderIdx})
		linkHeaderIdx = 1 + curMax
	}

	orderedRecordsByRowIdx := map[int][]string{}

	// TODO: take care of edge cases
	rowIdx := 0
	for _, record := range records {
		kvGroupKey := record[kvGroupHeaderIdx]

		// because we are modifying it
		if kvGroupKey == targetKvGroupKey {
			continue
		}

		rowIdxKey := rowIdx
		orderedRecordsByRowIdx[rowIdxKey] = record
		rowIdx++
	}

	var newRecord []string
	newRecord = make([]string, len(records[0]))

	newRecord[kvGroupHeaderIdx] = targetKvGroupKey
	newRecord[timeHeaderIdx] = time.Now().Format(time.DateTime)
	newRecord[linkHeaderIdx] = "https://uwaterloo-2.pagerduty.com/incidents"
	newRecord[detailsHeaderIdx] = newInvalidationsOutput.String()

	orderedRecordsByRowIdx[rowIdx] = newRecord

	csvBuffer, err := createCSVBuffer(orderedRecordsByRowIdx)
	if err != nil {
		return fmt.Errorf("error occurred while rendering the new CSV: %w", err)
	}

	if err := uploadToS3(s.s3Session, csvBuffer, bucketName, objectKey); err != nil {
		return fmt.Errorf("error occurred while updating new things to the CSV: %w", err)
	}
	return nil
}

func createCSVBuffer(data map[int][]string) (*bytes.Buffer, error) {
	sortedData := [][]string{}
	for i := 0; i < len(data); i++ {
		sortedData = append(sortedData, data[i])
	}

	// Create a buffer to hold the CSV data
	buffer := new(bytes.Buffer)
	writer := csv.NewWriter(buffer)

	// Write data to the CSV writer
	for _, record := range sortedData {
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}
	writer.Flush()

	return buffer, writer.Error()
}

func uploadToS3(s3session *session.Session, buffer *bytes.Buffer, bucketName, objectKey string) error {
	// Create an S3 service client
	svc := s3.New(s3session)

	// Upload the file
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(buffer.Bytes()),
		ContentType: aws.String("text/csv"),
	})

	return err
}
